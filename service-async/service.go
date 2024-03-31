package main

// @Title        service.go
// @Description  一个服务抽象的样例，支持自动向registry注册，并接入不同的服务实现
// @Create       dlchang (2024/03/14 16:30)

import (
	"context"
	"flag"
	"log"
	"net"
	"time"

	pb "github.com/bnkrr/iinode-demo/pb_autogen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	Network        string = "tcp"
	ServiceAddress string = "localhost:0"
	// RegistryAddress string = "localhost:48000"
)

var (
	registryAddress = flag.String("reg-addr", "localhost:48001", "Registry address in the format of host:port")
)

type ServiceMetadata struct {
	Name        string
	Version     string
	Concurrency int
	CallType    pb.CallType
}

// AsyncService 具体服务功能实现的接口定义
type AsyncService struct {
	pb.UnimplementedServiceServer // 来自grpc自动生成代码
	ServiceMetadata
	registryAddress *string           // registry的侦听地址
	netListener     *net.Listener     // 用于存储service的侦听
	registryClient  pb.RegistryClient // 用于调用registry的grpc client
}

// @title        CallAsync
// @description  启动一个服务的主函数
// @auth         dlchang (2024/03/14 16:30)
func (s *AsyncService) CallAsync(ctx context.Context, req *pb.ServiceCallRequest) (*pb.ServiceCallResponse, error) {
	go s.AsyncTask(context.Background(), req)
	return &pb.ServiceCallResponse{RequestId: req.RequestId, Output: "success"}, nil
}

func (s *AsyncService) AsyncTask(ctx context.Context, req *pb.ServiceCallRequest) {
	_, err := s.registryClient.SubmitResult(ctx, &pb.SubmitResultRequest{
		Name:      s.Name,
		RequestId: req.RequestId,
		End:       false,
		Output:    req.Input,
	})
	if err != nil {
		log.Printf("submit error %v\n", err)
	}
	_, err = s.registryClient.SubmitResult(ctx, &pb.SubmitResultRequest{
		Name:      s.Name,
		RequestId: req.RequestId,
		End:       true,
		Output:    req.Input + " new end",
	})
	if err != nil {
		log.Printf("submit error %v\n", err)
	}

	log.Printf("submitted, id:%v, input:%v\n", req.RequestId, req.Input)
}

// @title        Serve
// @description  启动一个服务的主函数
// @auth         dlchang (2024/03/14 16:30)
func (s *AsyncService) Serve() {
	listener, err := net.Listen(Network, ServiceAddress)
	if err != nil {
		log.Panicf("net.Listen err: %v", err)
	}
	s.netListener = &listener

	go s.RegisterRoutine()

	grpcServer := grpc.NewServer()
	pb.RegisterServiceServer(grpcServer, s)
	grpcServer.Serve(*s.netListener)
}

// @title        Register
// @description  向registry注册自己的端口等服务信息
// @auth         dlchang (2024/03/14 16:30)
func (s *AsyncService) Register() {
	if s.netListener == nil {
		return
	}
	port := (*s.netListener).Addr().(*net.TCPAddr).Port
	_, err := s.registryClient.RegisterService(context.Background(), &pb.RegisterServiceRequest{
		Name:        s.Name,
		Port:        int32(port),
		Version:     s.Version,
		Concurrency: int32(s.Concurrency),
		CallType:    s.CallType,
	})
	if err != nil {
		log.Printf("register err, %v", err)
	}
}

// @title        Register
// @description  周期性向registry注册
// @auth         dlchang (2024/03/14 16:30)
func (s *AsyncService) RegisterRoutine() {
	for {
		s.Register()
		time.Sleep(5 * time.Second)
	}
}

// @title        NewRegistryClient
// @description  新建registry grpc client
// @auth         dlchang (2024/03/27 16:47)
func NewRegistryClient(url string) (pb.RegistryClient, error) {
	conn, err := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	// defer conn.Close()
	return pb.NewRegistryClient(conn), nil
}

// @title        NewService
// @description  新建一个服务
// @auth         dlchang (2024/03/14 16:30)
func NewService(registryAddress *string, serviceMetadata *ServiceMetadata) (*AsyncService, error) {
	client, err := NewRegistryClient(*registryAddress)
	if err != nil {
		log.Panicf("net.Connect err, %v", err)
	}

	s := &AsyncService{
		ServiceMetadata: *serviceMetadata,
		registryAddress: registryAddress,
		registryClient:  client,
	}

	return s, nil
}

// @title        main
// @description  主函数，处理flag
// @auth         dlchang (2024/03/14 16:30)
func main() {
	flag.Parse()
	s, err := NewService(registryAddress, &ServiceMetadata{
		Name:        "AsyncService",
		Version:     "1.0.0",
		Concurrency: 10,
		CallType:    pb.CallType_ASYNC,
	})
	if err != nil {
		log.Panicf("new service err")
	}
	s.Serve()
}
