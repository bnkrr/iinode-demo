// 不同服务类型的实现

package main

import (
	"context"
	"log"

	pb "example.com/proto"
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
	return 1
}

// 服务的函数体，更改它实现具体的测量功能
// 输入、输出的数据结构可见proto/service.proto
// 输入结构体ServiceCallRequest有一个数据成员Input，是一个字符串（JSON）
// 输出结构体ServiceCallResponse有一个数据成员Output，是一个字符串（JSON）
func (s *EchoService) Call(ctx context.Context, req *pb.ServiceCallRequest) (*pb.ServiceCallResponse, error) {
	log.Printf("echo request: %s\n", req.Input)
	return &pb.ServiceCallResponse{Output: req.Input}, nil
}
