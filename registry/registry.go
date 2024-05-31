package main

// @Title        registry.go
// @Description  一个简单的供调试的注册服务器demo，只支持注册一个服务，并会周期性调用服务，打印输出
// @Create       dlchang (2024/03/14 16:30)

import (
	"context"
	"errors"
	"flag"
	"log"
	"net"
	"sync"

	pb "github.com/bnkrr/iinode-demo/pb_autogen"
	"google.golang.org/grpc"
)

const (
	Network string = "tcp"
)

var (
	registryAddress = flag.String("addr", "localhost:48001", "Listen address in the format of host:port")
	configPath      = flag.String("config", "config_example.json", "configuration path of registry")
)

type LocalService struct {
	Name        string      // 用于存储service的名称
	Port        int         // 用于存储service的端口
	CallType    pb.CallType // 用户存储service的调用类型
	Concurrency int         // 用户存储service的并发
}

func (s *LocalService) Same(other *LocalService) bool {
	return s.Name == other.Name &&
		s.Port == other.Port &&
		s.Concurrency == other.Concurrency &&
		s.CallType == other.CallType
}

// RegistryServer 供调试的注册服务器demo
type RegistryServer struct {
	pb.UnimplementedRegistryServer // 来自rpc生成代码
	config                         *ConfigRunners
	services                       sync.Map
	runners                        sync.Map
	ioManager                      *IOManager
}

// 删除runner信息
func (s *RegistryServer) CleanupRunner(serviceName string) {
	s.runners.Delete(serviceName)
}

// 查看service是否需要被运行，如果需要的话就运行
func (s *RegistryServer) CheckRunner(service *LocalService) {
	configService, ok := s.config.GetConfig(service)
	if !ok {
		return
	}

	oldRunner, ok := s.runners.Load(service.Name)
	if ok {
		r, ok := oldRunner.(*Runner)
		if !ok {
			log.Panic("err when loading runner")
		}
		r.Stop()
	}

	runner, err := NewRunner(service, configService)
	if err != nil {
		log.Panicf("err when creating runner, %v\n", err)
	}
	s.runners.Store(service.Name, runner)

	runner.StartWithIOManager(s.ioManager)
	runner.Serve()

	// 在服务关闭后删除runner
	s.CleanupRunner(service.Name)
}

// 检查一个service是否已存在
func (s *RegistryServer) ServiceExists(service *LocalService) (bool, error) {
	oldServiceAny, ok := s.services.Load(service.Name)
	if !ok {
		return false, nil
	}
	oldService, ok := oldServiceAny.(*LocalService)
	if !ok {
		return false, errors.New("load service err")
	}
	return oldService.Same(service), nil
}

// @title        RegisterService
// @description  rpc接口实现，注册一个服务
// @auth         dlchang (2024/03/14 16:30)
// @param        ctx   context.Context   上下文对象
// @param        req   *pb.RegisterServiceRequest   服务注册请求的rpc生成类型，内含服务各项信息，比如端口、是否是流调用等等。
// @return       *pb.RegisterServiceResponse   服务注册响应的rpc生成类型
func (s *RegistryServer) RegisterService(ctx context.Context, req *pb.RegisterServiceRequest) (*pb.RegisterServiceResponse, error) {
	log.Printf("%s listen at localhost:%d\n", req.Name, req.Port)
	newService := &LocalService{
		Name:        req.Name,
		Port:        int(req.Port),
		Concurrency: int(req.Concurrency),
		CallType:    req.CallType,
	}

	// 检查信息是否更改，如果是新的服务则需要运行
	msg := "kept"
	exists, err := s.ServiceExists(newService)
	if err != nil {
		return nil, err
	}
	if !exists {
		msg = "updated"
		go s.CheckRunner(newService)
		s.services.Store(req.Name, newService)
	}

	return &pb.RegisterServiceResponse{Message: msg}, nil
}

func (s *RegistryServer) SubmitResult(ctx context.Context, req *pb.SubmitResultRequest) (*pb.SubmitResultResponse, error) {
	runner, ok := s.runners.Load(req.Name)
	if !ok {
		return nil, errors.New("runner not found")
	}
	r, ok := runner.(*Runner)
	if !ok {
		return nil, errors.New("runner loading error")
	}
	r.ReceiveResult(req)
	return &pb.SubmitResultResponse{Message: "success"}, nil
}

func (s *RegistryServer) Serve() {
	// rabbitmq
	iomgr, err := NewIOManager(s.config)
	if err != nil {
		log.Panicf("IO error: %v\n", err)
	}
	s.ioManager = iomgr

	// grpc
	listener, err := net.Listen(Network, *registryAddress)
	if err != nil {
		log.Panicf("net.Listen err: %v", err)
	}
	log.Printf("registry listen at %s\n", *registryAddress)

	grpcServer := grpc.NewServer()
	pb.RegisterRegistryServer(grpcServer, s)
	grpcServer.Serve(listener)
}

func NewRegistryServer() *RegistryServer {
	config := &ConfigRunners{}
	config.LoadConfigFromFile(*configPath)
	log.Printf("config loaded, allow-no-name mode: %v\n", config.RabbitMQ.AllowNoName)

	return &RegistryServer{
		config:    config,
		services:  sync.Map{},
		runners:   sync.Map{},
		ioManager: nil,
	}
}

// @title        main
// @description  运行RegistryServer的grpc服务器
// @auth         dlchang (2024/03/14 16:30)
func main() {
	flag.Parse()

	r := NewRegistryServer()
	r.Serve()

}
