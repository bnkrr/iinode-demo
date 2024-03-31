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

// ServiceHandler 具体服务功能实现的接口定义
type ServiceHandler interface {
	Name() string                                                                       // 服务名称
	Call(context.Context, *pb.ServiceCallRequest) (*pb.ServiceCallResponse, error)      // 服务的普通调用实现
	CallStream(*pb.ServiceCallRequest, pb.Service_CallStreamServer) error               // 服务的流调用实现
	CallAsync(context.Context, *pb.ServiceCallRequest) (*pb.ServiceCallResponse, error) // 服务的异步调用实现
	Version() string                                                                    // 服务版本
	Concurrency() int32                                                                 // 服务支持的并发
	CallType() pb.CallType                                                              // 服务是否实现了流调用
}

// BaseService 具体服务功能实现的接口定义
type BaseService struct {
	pb.UnimplementedServiceServer                   // 来自grpc自动生成代码
	service                       ServiceHandler    // 实现的服务接口
	registryAddress               *string           // registry的侦听地址
	netListener                   *net.Listener     // 用于存储service的侦听
	registryClient                pb.RegistryClient // 用于调用registry的grpc client
}

// @title        Call
// @description  rpc接口实现，服务普通调用，用于实现一次性输入输出的服务，如测量一个IP地址是否有指定漏洞
// @auth         dlchang (2024/03/14 16:30)
// @param        ctx   context.Context   上下文对象
// @param        req   *pb.ServiceCallRequest   调用请求的rpc生成类型，内含输入Input字符串。
// @return       *pb.ServiceCallResponse  调用响应的rpc生成类型，内含输出Output字符串
// @return       error   错误
func (s *BaseService) Call(ctx context.Context, req *pb.ServiceCallRequest) (*pb.ServiceCallResponse, error) {
	return s.service.Call(ctx, req)
}

// @title        CallStream
// @description  rpc接口实现，服务流调用，用于实现返回一个流的服务，如测量一段IP地址是否有活跃端口
// @auth         dlchang (2024/03/14 16:30)
// @param        req   *pb.ServiceCallRequest   调用请求的rpc生成类型，内含输入Input字符串。
// @param        stream pb.Service_CallStreamServer  调用响应的rpc生成类型，是一个流
// @return       error   错误
func (s *BaseService) CallStream(req *pb.ServiceCallRequest, stream pb.Service_CallStreamServer) error {
	return s.service.CallStream(req, stream)
}

func (s *BaseService) CallAsync(ctx context.Context, req *pb.ServiceCallRequest) (*pb.ServiceCallResponse, error) {
	return s.service.CallAsync(ctx, req)
}

// @title        Serve
// @description  启动一个服务的主函数
// @auth         dlchang (2024/03/14 16:30)
func (s *BaseService) Serve() {
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
func (s *BaseService) Register() {
	conn, err := grpc.Dial(*s.registryAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Panicf("net.Connect err: %v", err)
	}
	defer conn.Close()
	s.registryClient = pb.NewRegistryClient(conn)

	if s.netListener == nil {
		return
	}
	port := (*s.netListener).Addr().(*net.TCPAddr).Port
	_, err = s.registryClient.RegisterService(context.Background(), &pb.RegisterServiceRequest{
		Name:        s.service.Name(),
		Port:        int32(port),
		Version:     s.service.Version(),
		Concurrency: s.service.Concurrency(),
		CallType:    s.service.CallType(),
	})
	if err != nil {
		log.Printf("register err, %v", err)
	}
}

// @title        Register
// @description  周期性向registry注册
// @auth         dlchang (2024/03/14 16:30)
func (s *BaseService) RegisterRoutine() {
	for {
		s.Register()
		time.Sleep(5 * time.Second)
	}
}

// @title        NewService
// @description  新建一个服务
// @auth         dlchang (2024/03/14 16:30)
func NewService(registryAddress *string, service ServiceHandler) (*BaseService, error) {
	s := &BaseService{registryAddress: registryAddress, service: service}

	return s, nil
}

// @title        main
// @description  主函数，处理flag
// @auth         dlchang (2024/03/14 16:30)
func main() {
	flag.Parse()
	s, err := NewService(registryAddress, &EchoService{}) // 改动此处服务类型
	// s, err := NewService(registryAddress, &AsyncService{}) // 改动此处服务类型
	// s, err := NewService(registryAddress, &StreamService{}) // 改动此处服务类型
	if err != nil {
		log.Panicf("new service err")
	}
	s.Serve()
}
