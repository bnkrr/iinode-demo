// 不同服务类型的实现

package main

import (
	"context"
	"errors"
	"log"
	"time"

	pb "github.com/bnkrr/iinode-demo/pb_autogen"
)

type EchoService struct{}

// 服务名称
func (s *EchoService) Name() string {
	return "EchoService"
}

// 服务版本
func (s *EchoService) Version() string {
	return "1.0.0"
}

// 服务运行的并发
func (s *EchoService) Concurrency() int32 {
	return 10
}

// 返回是否是一个流
func (s *EchoService) CallType() pb.CallType {
	return pb.CallType_NORMAL
}

// 服务的函数体，更改它实现具体的测量功能
// 输入、输出的数据结构可见proto/service.proto
// 输入结构体ServiceCallRequest有一个数据成员Input，是一个字符串（JSON）
// 输出结构体ServiceCallResponse有一个数据成员Output，是一个字符串（JSON）
func (s *EchoService) Call(ctx context.Context, req *pb.ServiceCallRequest) (*pb.ServiceCallResponse, error) {
	log.Printf("echo request: %s\n", req.Input)
	time.Sleep(time.Duration(567) * time.Microsecond)
	return &pb.ServiceCallResponse{Output: req.Input}, nil
}

// 直接调用服务，无需实现stream方法
func (s *EchoService) CallStream(req *pb.ServiceCallRequest, stream pb.Service_CallStreamServer) error {
	return errors.New("not implemented")
}

// 直接调用服务，无需实现async方法
func (s *EchoService) CallAsync(ctx context.Context, req *pb.ServiceCallRequest) (*pb.ServiceCallResponse, error) {
	return nil, errors.New("not implemented")
}
