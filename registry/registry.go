package main

// @Title        registry.go
// @Description  一个简单的供调试的注册服务器demo，只支持注册一个服务，并会周期性调用服务，打印输出
// @Create       dlchang (2024/03/14 16:30)

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	pb "github.com/bnkrr/iinode-demo/pb_autogen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	Network string = "tcp"
)

var (
	registryAddress = flag.String("addr", "localhost:48001", "Listen address in the format of host:port")
)

// RegistryServer 供调试的注册服务器demo
type RegistryServer struct {
	pb.UnimplementedRegistryServer      // 来自rpc生成代码
	servicePort                    int  // 用于存储service的端口
	returnStream                   bool // 用户存储service是否是流服务
}

// @title        RegisterService
// @description  rpc接口实现，注册一个服务
// @auth         dlchang (2024/03/14 16:30)
// @param        ctx   context.Context   上下文对象
// @param        req   *pb.RegisterServiceRequest   服务注册请求的rpc生成类型，内含服务各项信息，比如端口、是否是流调用等等。
// @return       *pb.RegisterServiceResponse   服务注册响应的rpc生成类型
func (s *RegistryServer) RegisterService(ctx context.Context, req *pb.RegisterServiceRequest) (*pb.RegisterServiceResponse, error) {
	log.Printf("%s listen at localhost:%d\n", req.Name, req.Port)
	s.servicePort = int(req.Port)
	s.returnStream = req.ReturnStream
	return &pb.RegisterServiceResponse{Message: "success"}, nil
}

// @title        Run
// @description  调用注册service提供的普通函数调用接口
// @auth         dlchang (2024/03/14 16:30)
func (s *RegistryServer) Run() {
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", s.servicePort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Panicf("dial net.Connect err: %v", err)
	}
	defer conn.Close()
	runnerClient := pb.NewServiceClient(conn)
	resp, err := runnerClient.Call(context.Background(), &pb.ServiceCallRequest{Input: "1.2.3.4"})
	if err != nil {
		log.Printf("call err, try again later: %v\n", err)
	} else {
		log.Printf("output: %s\n", resp.Output)
	}
}

// @title        RunStream
// @description  调用注册service提供的返回流的调用接口
// @auth         dlchang (2024/03/14 16:30)
func (s *RegistryServer) RunStream() {
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", s.servicePort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Panicf("dial net.Connect err: %v", err)
	}
	defer conn.Close()
	runnerClient := pb.NewServiceClient(conn)
	stream, err := runnerClient.CallStream(context.Background(), &pb.ServiceCallRequest{Input: "something"})
	if err != nil {
		log.Printf("call err, try again later: %v\n", err)
	} else {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("%v.CallStream, %v", runnerClient, err)
			}
			log.Printf("output: %s\n", resp.Output)
		}
	}
}

// @title        RunRoutine
// @description  根据service提供的调用类型，定期调用注册service提供的调用接口
// @auth         dlchang (2024/03/14 16:30)
func (s *RegistryServer) RunRoutine() {
	for {
		if s.servicePort < 0 {
			continue
		}
		if s.returnStream {
			s.RunStream()
		} else {
			s.Run()
		}
		time.Sleep(3 * time.Second)
	}
}

// @title        main
// @description  运行RegistryServer的grpc服务器
// @auth         dlchang (2024/03/14 16:30)
func main() {
	flag.Parse()
	listener, err := net.Listen(Network, *registryAddress)
	if err != nil {
		log.Panicf("net.Listen err: %v", err)
	}
	log.Printf("registry listen at %s\n", *registryAddress)

	grpcServer := grpc.NewServer()
	regServer := RegistryServer{servicePort: -1}
	go regServer.RunRoutine()
	pb.RegisterRegistryServer(grpcServer, &regServer)
	grpcServer.Serve(listener)
}
