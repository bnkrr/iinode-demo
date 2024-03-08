package main

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

type ServiceHandler interface {
	Name() string
	Call(context.Context, *pb.ServiceCallRequest) (*pb.ServiceCallResponse, error)
	CallStream(*pb.ServiceCallRequest, pb.Service_CallStreamServer) error
	Version() string
	Concurrency() int32
	ReturnStream() bool
}

type BaseService struct {
	pb.UnimplementedServiceServer
	service         ServiceHandler
	registryAddress *string
	netListener     *net.Listener
	registryClient  pb.RegistryClient
}

func (s *BaseService) Call(ctx context.Context, req *pb.ServiceCallRequest) (*pb.ServiceCallResponse, error) {
	return s.service.Call(ctx, req)
}

func (s *BaseService) CallStream(req *pb.ServiceCallRequest, stream pb.Service_CallStreamServer) error {
	return s.service.CallStream(req, stream)
}

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
		Name:         s.service.Name(),
		Port:         int32(port),
		Version:      s.service.Version(),
		Concurrency:  s.service.Concurrency(),
		ReturnStream: s.service.ReturnStream(),
	})
	if err != nil {
		log.Printf("register err, %v", err)
	}
}

func (s *BaseService) RegisterRoutine() {
	for {
		s.Register()
		time.Sleep(5 * time.Second)
	}
}

func NewService(registryAddress *string, service ServiceHandler) (*BaseService, error) {
	s := &BaseService{registryAddress: registryAddress, service: service}

	return s, nil
}

func main() {
	flag.Parse()
	s, err := NewService(registryAddress, &EchoService{}) // 改动此处服务类型
	// s, err := NewService(registryAddress, &StreamService{}) // 改动此处服务类型
	if err != nil {
		log.Panicf("new service err")
	}
	s.Serve()
}
