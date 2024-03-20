package main

// @Title        registry.go
// @Description  一个简单的供调试的注册服务器demo，只支持注册一个服务，并会周期性调用服务，打印输出
// @Create       dlchang (2024/03/14 16:30)

import (
	"context"
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
	Name         string // 用于存储service的名称
	Port         int    // 用于存储service的端口
	ReturnStream bool   // 用户存储service是否是流服务
	Concurrency  int    // 用户存储service的并发
}

func (s *LocalService) Same(other *LocalService) bool {
	return s.Name == other.Name &&
		s.Port == other.Port &&
		s.Concurrency == other.Concurrency &&
		s.ReturnStream == other.ReturnStream
}

// RegistryServer 供调试的注册服务器demo
type RegistryServer struct {
	pb.UnimplementedRegistryServer // 来自rpc生成代码
	config                         *ConfigRunners
	services                       map[string](*LocalService)
	runners                        map[string](*Runner)
	registerLock                   sync.Mutex
}

// 查看service是否需要被运行，如果需要的话就运行
func (s *RegistryServer) CheckRunner(service *LocalService) {
	configService, ok := s.config.GetConfig(service)
	if !ok {
		return
	}
	oldRunner, ok := s.runners[service.Name]
	if ok {
		oldRunner.Stop()
	}
	runner, err := NewRunner(service, configService)
	if err != nil {
		log.Panicf("err when creating runner, %v\n", err)
	}
	s.runners[service.Name] = runner
	runner.Start()
	runner.Serve()
}

// @title        RegisterService
// @description  rpc接口实现，注册一个服务
// @auth         dlchang (2024/03/14 16:30)
// @param        ctx   context.Context   上下文对象
// @param        req   *pb.RegisterServiceRequest   服务注册请求的rpc生成类型，内含服务各项信息，比如端口、是否是流调用等等。
// @return       *pb.RegisterServiceResponse   服务注册响应的rpc生成类型
func (s *RegistryServer) RegisterService(ctx context.Context, req *pb.RegisterServiceRequest) (*pb.RegisterServiceResponse, error) {
	s.registerLock.Lock()
	defer s.registerLock.Unlock()

	log.Printf("%s listen at localhost:%d\n", req.Name, req.Port)
	newService := &LocalService{
		Name:         req.Name,
		Port:         int(req.Port),
		Concurrency:  int(req.Concurrency),
		ReturnStream: req.ReturnStream,
	}

	msg := "kept"
	oldService, ok := s.services[req.Name]
	if !(ok && oldService.Same(newService)) {
		msg = "updated"
		go s.CheckRunner(newService)
	}
	s.services[req.Name] = newService

	return &pb.RegisterServiceResponse{Message: msg}, nil
}

func (s *RegistryServer) Serve() {
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

	return &RegistryServer{
		config:       config,
		services:     make(map[string]*LocalService),
		runners:      make(map[string]*Runner),
		registerLock: sync.Mutex{},
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
