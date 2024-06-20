package main

// @Title        registry.go
// @Description  一个简单的供调试的注册服务器demo，只支持注册一个服务，并会周期性调用服务，打印输出
// @Create       dlchang (2024/03/14 16:30)

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	pb "github.com/bnkrr/iinode-demo/pb_autogen"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"google.golang.org/grpc"
)

const (
	Network string = "tcp"
)

type LocalServiceRegisterStatusType int

const (
	LocalServiceRegisterStatus_NEW    LocalServiceRegisterStatusType = iota // 新服务
	LocalServiceRegisterStatus_RENEW                                        // 旧服务续租
	LocalServiceRegisterStatus_UPDATE                                       // 更新服务
	LocalServiceRegisterStatus_IDLE                                         // 什么都不做
	LocalServiceRegisterStatus_STOP                                         // 停止现有服务
	LocalServiceRegisterStatus_ERROR                                        // 发生错误
	LocalServiceRegisterStatus_OTHER                                        // 其他情况
)

func GetTimestamp() int64 {
	return time.Now().UnixMilli()
}

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
	config                         *ConfigRegistry
	services                       sync.Map
	runners                        sync.Map
	ioManager                      *IOManager
	etcdManager                    *EtcdManager
}

// 删除runner/service信息，如果不存在，不造成影响
func (s *RegistryServer) CleanupRunner(serviceName string) {
	s.runners.Delete(serviceName)
	s.services.Delete(serviceName)
}

// 查看service是否需要被运行，如果需要的话就运行
func (s *RegistryServer) CheckRunner(service *LocalService) {
	configRunner, ok := s.config.GetConfigRunner(service.Name)
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

	runner, err := NewRunner(service, configRunner)
	if err != nil {
		log.Panicf("err when creating runner, %v\n", err)
	}
	s.runners.Store(service.Name, runner)

	runner.StartWithIOManager(s.ioManager)
	runner.Serve()

	// 在服务关闭后删除runner和service
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
	log.Printf("%s(localhost:%d) try to register\n", req.Name, req.Port)
	newService := &LocalService{
		Name:        req.Name,
		Port:        int(req.Port),
		Concurrency: int(req.Concurrency),
		CallType:    req.CallType,
	}

	// 检查信息是否更改，如果是新的服务则需要运行
	s.services.Store(req.Name, newService)
	status, err := s.UpdateLocalService(newService)
	if err != nil {
		return nil, err
	}
	switch status {
	case LocalServiceRegisterStatus_UPDATE:
		return &pb.RegisterServiceResponse{Message: "updated"}, nil
	case LocalServiceRegisterStatus_STOP:
		return &pb.RegisterServiceResponse{Message: "stopped"}, nil
	}
	return &pb.RegisterServiceResponse{Message: "kept"}, nil
}

func (s *RegistryServer) SubmitResult(ctx context.Context, req *pb.SubmitResultRequest) (*pb.SubmitResultResponse, error) {
	runner, ok := s.runners.Load(req.Name)
	if !ok {
		return nil, errors.New("service not found")
	}
	r, ok := runner.(*Runner)
	if !ok {
		return nil, errors.New("service loading error")
	}
	r.ReceiveResult(req)
	return &pb.SubmitResultResponse{Message: "success"}, nil
}

func (s *RegistryServer) CreateRunner(service *LocalService, configRunner *ConfigRunner) {
	// 新建runner
	runner, err := NewRunner(service, configRunner)
	if err != nil {
		log.Panicf("err when creating runner, %v\n", err)
	}
	s.runners.Store(service.Name, runner)

	runner.StartWithIOManager(s.ioManager)
	runner.Serve()

	// 在服务关闭后删除runner和service
	s.CleanupRunner(service.Name)
}

// 依照旧service对runner进行update
// 1.无新config runner
// 1.0.无runner，无动作
// 1.1.有runner，结束现有runner
// 2.有新config runner
// 2.1.无runner，新建runner运行
// 2.2.有runner
// 2.2.1. config runner未更新，无动作
// 2.2.2. config runner已更新，停止现有runner，运行新runner
func (s *RegistryServer) UpdateLocalService(service *LocalService) (LocalServiceRegisterStatusType, error) {
	configRunner, configRunnerFound := s.config.GetConfigRunner(service.Name)
	oldRunner, runnerFound := s.runners.Load(service.Name)

	// 1.无新config runner
	if !configRunnerFound {
		// 1.1.有runner，结束现有runner
		if runnerFound {
			r, ok := oldRunner.(*Runner)
			if ok {
				r.Stop()
				return LocalServiceRegisterStatus_STOP, nil
			}
		}
		return LocalServiceRegisterStatus_OTHER, nil
	}

	// 2.有新config runner
	// 2.1.无runner，新建runner运行
	if !runnerFound {
		go s.CreateRunner(service, configRunner)
		return LocalServiceRegisterStatus_IDLE, nil
	}

	// 2.2.有runner
	r, ok := oldRunner.(*Runner)
	if !ok {
		return LocalServiceRegisterStatus_ERROR, errors.New("load new runner error")
	}

	switch s.config.MatchConfigRunner(r.config) {
	// 2.2.0. config runner找不到（不可能发生）
	case ConfigRunner_NOTFOUND:
		r.Stop()
		return LocalServiceRegisterStatus_STOP, nil
	// 2.2.1. config runner未更新，无动作
	case ConfigRunner_SAME:
		return LocalServiceRegisterStatus_IDLE, nil
	// 2.2.2. config runner已更新，停止现有runner，运行新runner
	case ConfigRunner_DIFFERENT:
		r.Stop()
		go s.CreateRunner(service, configRunner)
	}
	return LocalServiceRegisterStatus_UPDATE, nil
}

func (s *RegistryServer) UpdateConfig(cfgbytes []byte) error {
	cfg := ConfigRegistry{}
	err := cfg.LoadConfigFromBytes(cfgbytes)
	if err != nil {
		return err
	}

	// fmt.Println(cfg.RabbitMQ.Url == "")
	s.config = &cfg
	// TODO: 处理新MQ问题
	// 初始化rabbitmq
	if s.config.RabbitMQ.Url != "" && s.ioManager == nil {
		iomgr, err := NewIOManager(s.config)
		if err != nil {
			return err
		}
		s.ioManager = iomgr
		fmt.Printf("mq -> %s", cfg.RabbitMQ.Url)
	}

	// 处理新service问题
	// 1.更换config
	// 2.依照本地存在的service对runner进行update
	s.services.Range(func(name any, value any) bool {
		service, ok := value.(*LocalService)
		if !ok {
			return true
		}
		s.UpdateLocalService(service)
		return true
	})

	return nil
}

func (s *RegistryServer) Serve() {
	ctx, cancel := context.WithCancel(context.TODO())
	// etcd
	emgr, err := NewEtcdManager(
		s.config.Etcd.Url,
		s.config.Etcd.User,
		s.config.Etcd.Password,
		s.config.Etcd.Prefix,
	)
	if err != nil {
		log.Panicf("connect to etcd error: %v\n", err)
	}

	s.etcdManager = emgr
	s.etcdManager.RefreshMetadata(ctx)
	ttlCh, err := s.etcdManager.Register(ctx)
	if err != nil {
		log.Panicf("register error: %v\n", err)
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ttlCh:
				// log.Printf("ttl, %v", resp.TTL)
			}
		}
	}()

	// 初始化并监视更新config
	cfgbytes := s.etcdManager.GetConfigNoError(ctx)
	err = s.UpdateConfig(cfgbytes)
	if err != nil {
		log.Panicf("init error: %v\n", err)
	}

	// 监视config变化
	cfgCh := s.etcdManager.WatchConfig(ctx)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case resp := <-cfgCh:
				for _, event := range resp.Events {
					switch event.Type {
					case mvccpb.PUT:
						s.UpdateConfig(event.Kv.Value)
					}
				}
			}
		}
	}()

	// grpc
	listener, err := net.Listen(Network, *registryAddress)
	if err != nil {
		log.Panicf("net.Listen err: %v", err)
	}
	log.Printf("registry listen at %s\n", *registryAddress)

	grpcServer := grpc.NewServer()
	pb.RegisterRegistryServer(grpcServer, s)
	grpcServer.Serve(listener)

	cancel()
}

func NewRegistryServer() *RegistryServer {
	config := &ConfigRegistry{}
	config.LoadConfigFromFile(*configPath)
	//log.Printf("config loaded, allow-no-name mode: %v\n", config.RabbitMQ.AllowNoName)

	return &RegistryServer{
		config:      config,
		services:    sync.Map{},
		runners:     sync.Map{},
		ioManager:   nil,
		etcdManager: nil,
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
