package main

import (
	"context"
	"flag"
	"fmt"
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

type RegistryServer struct {
	pb.UnimplementedRegistryServer
	servicePort int
}

func (s *RegistryServer) RegisterService(ctx context.Context, req *pb.RegisterServiceRequest) (*pb.RegisterServiceResponse, error) {
	log.Printf("%s listen at localhost:%d\n", req.Name, req.Port)
	s.servicePort = int(req.Port)
	return &pb.RegisterServiceResponse{Message: "success"}, nil
}

func (s *RegistryServer) Run() {
	for {
		if s.servicePort < 0 {
			continue
		}

		conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", s.servicePort), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Panicf("dial net.Connect err: %v", err)
		}
		defer conn.Close()
		runnerClient := pb.NewServiceClient(conn)
		resp, err := runnerClient.Call(context.Background(), &pb.ServiceCallRequest{Input: "something"})
		if err != nil {
			log.Printf("call err, try again later: %v\n", err)
		} else {
			log.Printf("output: %s\n", resp.Output)
		}
		time.Sleep(3 * time.Second)
	}

}

func main() {
	flag.Parse()
	listener, err := net.Listen(Network, *registryAddress)
	if err != nil {
		log.Panicf("net.Listen err: %v", err)
	}
	log.Printf("registry listen at %s\n", *registryAddress)

	grpcServer := grpc.NewServer()
	regServer := RegistryServer{servicePort: -1}
	go regServer.Run()
	pb.RegisterRegistryServer(grpcServer, &regServer)
	grpcServer.Serve(listener)
}
