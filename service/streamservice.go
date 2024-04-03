// 不同服务类型的实现

package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	pb "github.com/bnkrr/iinode-demo/pb_autogen"
)

type StreamService struct{}

// 服务名称
func (s *StreamService) Name() string {
	return "StreamService"
}

// 服务版本
func (s *StreamService) Version() string {
	return "1.0.0"
}

// 服务运行的并发
func (s *StreamService) Concurrency() int32 {
	return 1
}

// 返回是否是一个流
func (s *StreamService) CallType() pb.CallType {
	return pb.CallType_STREAM
}

// 直接调用函数，实现流方法则无需实现本方法
func (s *StreamService) Call(ctx context.Context, req *pb.ServiceCallRequest) (*pb.ServiceCallResponse, error) {
	return nil, errors.New("not implemented")
}

// 服务的函数体，更改它实现具体的测量功能（返回一个流）
// 输入、输出的数据结构可见proto/service.proto
// 输入结构体ServiceCallRequest有一个数据成员Input，是一个字符串（JSON）
// 输出由形参传入，是一个流，可以调用Send()方法，发送一个ServiceCallResponse结构体
// 发送完成后直接返回即可
func (s *StreamService) CallStream(req *pb.ServiceCallRequest, stream pb.Service_CallStreamServer) error {
	log.Printf("stream request: %s\n", req.Input)
	for i := 1; i <= 3; i++ {
		outStr := fmt.Sprintf("result %d for input %s", i, req.Input)
		if err := stream.Send(&pb.ServiceCallResponse{Output: outStr}); err != nil {
			return err
		}
	}
	return nil
}

// 直接调用服务，无需实现async方法
func (s *StreamService) CallAsync(ctx context.Context, req *pb.ServiceCallRequest) (*pb.ServiceCallResponse, error) {
	return nil, errors.New("not implemented")
}
