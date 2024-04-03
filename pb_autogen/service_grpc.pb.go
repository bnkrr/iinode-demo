// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.3
// source: service.proto

package pb_autogen

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Service_Call_FullMethodName       = "/pb_autogen.Service/Call"
	Service_CallStream_FullMethodName = "/pb_autogen.Service/CallStream"
	Service_CallAsync_FullMethodName  = "/pb_autogen.Service/CallAsync"
)

// ServiceClient is the client API for Service service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ServiceClient interface {
	Call(ctx context.Context, in *ServiceCallRequest, opts ...grpc.CallOption) (*ServiceCallResponse, error)
	CallStream(ctx context.Context, in *ServiceCallRequest, opts ...grpc.CallOption) (Service_CallStreamClient, error)
	CallAsync(ctx context.Context, in *ServiceCallRequest, opts ...grpc.CallOption) (*ServiceCallResponse, error)
}

type serviceClient struct {
	cc grpc.ClientConnInterface
}

func NewServiceClient(cc grpc.ClientConnInterface) ServiceClient {
	return &serviceClient{cc}
}

func (c *serviceClient) Call(ctx context.Context, in *ServiceCallRequest, opts ...grpc.CallOption) (*ServiceCallResponse, error) {
	out := new(ServiceCallResponse)
	err := c.cc.Invoke(ctx, Service_Call_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) CallStream(ctx context.Context, in *ServiceCallRequest, opts ...grpc.CallOption) (Service_CallStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &Service_ServiceDesc.Streams[0], Service_CallStream_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &serviceCallStreamClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Service_CallStreamClient interface {
	Recv() (*ServiceCallResponse, error)
	grpc.ClientStream
}

type serviceCallStreamClient struct {
	grpc.ClientStream
}

func (x *serviceCallStreamClient) Recv() (*ServiceCallResponse, error) {
	m := new(ServiceCallResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *serviceClient) CallAsync(ctx context.Context, in *ServiceCallRequest, opts ...grpc.CallOption) (*ServiceCallResponse, error) {
	out := new(ServiceCallResponse)
	err := c.cc.Invoke(ctx, Service_CallAsync_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ServiceServer is the server API for Service service.
// All implementations must embed UnimplementedServiceServer
// for forward compatibility
type ServiceServer interface {
	Call(context.Context, *ServiceCallRequest) (*ServiceCallResponse, error)
	CallStream(*ServiceCallRequest, Service_CallStreamServer) error
	CallAsync(context.Context, *ServiceCallRequest) (*ServiceCallResponse, error)
	mustEmbedUnimplementedServiceServer()
}

// UnimplementedServiceServer must be embedded to have forward compatible implementations.
type UnimplementedServiceServer struct {
}

func (UnimplementedServiceServer) Call(context.Context, *ServiceCallRequest) (*ServiceCallResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Call not implemented")
}
func (UnimplementedServiceServer) CallStream(*ServiceCallRequest, Service_CallStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method CallStream not implemented")
}
func (UnimplementedServiceServer) CallAsync(context.Context, *ServiceCallRequest) (*ServiceCallResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CallAsync not implemented")
}
func (UnimplementedServiceServer) mustEmbedUnimplementedServiceServer() {}

// UnsafeServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ServiceServer will
// result in compilation errors.
type UnsafeServiceServer interface {
	mustEmbedUnimplementedServiceServer()
}

func RegisterServiceServer(s grpc.ServiceRegistrar, srv ServiceServer) {
	s.RegisterService(&Service_ServiceDesc, srv)
}

func _Service_Call_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ServiceCallRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceServer).Call(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Service_Call_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServiceServer).Call(ctx, req.(*ServiceCallRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Service_CallStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ServiceCallRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ServiceServer).CallStream(m, &serviceCallStreamServer{stream})
}

type Service_CallStreamServer interface {
	Send(*ServiceCallResponse) error
	grpc.ServerStream
}

type serviceCallStreamServer struct {
	grpc.ServerStream
}

func (x *serviceCallStreamServer) Send(m *ServiceCallResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Service_CallAsync_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ServiceCallRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceServer).CallAsync(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Service_CallAsync_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServiceServer).CallAsync(ctx, req.(*ServiceCallRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Service_ServiceDesc is the grpc.ServiceDesc for Service service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Service_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pb_autogen.Service",
	HandlerType: (*ServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Call",
			Handler:    _Service_Call_Handler,
		},
		{
			MethodName: "CallAsync",
			Handler:    _Service_CallAsync_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "CallStream",
			Handler:       _Service_CallStream_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "service.proto",
}
