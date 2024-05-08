// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: spark/v1/spark_service.proto

package sparkv1

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

// SparkServiceClient is the client API for SparkService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SparkServiceClient interface {
	// Run a job
	RunJob(ctx context.Context, in *RunJobRequest, opts ...grpc.CallOption) (*RunJobResponse, error)
	// Stream job logs
	StreamJobLogs(ctx context.Context, in *StreamJobLogsRequest, opts ...grpc.CallOption) (SparkService_StreamJobLogsClient, error)
	// Ping the server
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
}

type sparkServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewSparkServiceClient(cc grpc.ClientConnInterface) SparkServiceClient {
	return &sparkServiceClient{cc}
}

func (c *sparkServiceClient) RunJob(ctx context.Context, in *RunJobRequest, opts ...grpc.CallOption) (*RunJobResponse, error) {
	out := new(RunJobResponse)
	err := c.cc.Invoke(ctx, "/spark.v1.SparkService/RunJob", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sparkServiceClient) StreamJobLogs(ctx context.Context, in *StreamJobLogsRequest, opts ...grpc.CallOption) (SparkService_StreamJobLogsClient, error) {
	stream, err := c.cc.NewStream(ctx, &SparkService_ServiceDesc.Streams[0], "/spark.v1.SparkService/StreamJobLogs", opts...)
	if err != nil {
		return nil, err
	}
	x := &sparkServiceStreamJobLogsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type SparkService_StreamJobLogsClient interface {
	Recv() (*StreamJobLogsResponse, error)
	grpc.ClientStream
}

type sparkServiceStreamJobLogsClient struct {
	grpc.ClientStream
}

func (x *sparkServiceStreamJobLogsClient) Recv() (*StreamJobLogsResponse, error) {
	m := new(StreamJobLogsResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *sparkServiceClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	out := new(PingResponse)
	err := c.cc.Invoke(ctx, "/spark.v1.SparkService/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SparkServiceServer is the server API for SparkService service.
// All implementations should embed UnimplementedSparkServiceServer
// for forward compatibility
type SparkServiceServer interface {
	// Run a job
	RunJob(context.Context, *RunJobRequest) (*RunJobResponse, error)
	// Stream job logs
	StreamJobLogs(*StreamJobLogsRequest, SparkService_StreamJobLogsServer) error
	// Ping the server
	Ping(context.Context, *PingRequest) (*PingResponse, error)
}

// UnimplementedSparkServiceServer should be embedded to have forward compatible implementations.
type UnimplementedSparkServiceServer struct {
}

func (UnimplementedSparkServiceServer) RunJob(context.Context, *RunJobRequest) (*RunJobResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RunJob not implemented")
}
func (UnimplementedSparkServiceServer) StreamJobLogs(*StreamJobLogsRequest, SparkService_StreamJobLogsServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamJobLogs not implemented")
}
func (UnimplementedSparkServiceServer) Ping(context.Context, *PingRequest) (*PingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}

// UnsafeSparkServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SparkServiceServer will
// result in compilation errors.
type UnsafeSparkServiceServer interface {
	mustEmbedUnimplementedSparkServiceServer()
}

func RegisterSparkServiceServer(s grpc.ServiceRegistrar, srv SparkServiceServer) {
	s.RegisterService(&SparkService_ServiceDesc, srv)
}

func _SparkService_RunJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RunJobRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SparkServiceServer).RunJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/spark.v1.SparkService/RunJob",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SparkServiceServer).RunJob(ctx, req.(*RunJobRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SparkService_StreamJobLogs_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StreamJobLogsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SparkServiceServer).StreamJobLogs(m, &sparkServiceStreamJobLogsServer{stream})
}

type SparkService_StreamJobLogsServer interface {
	Send(*StreamJobLogsResponse) error
	grpc.ServerStream
}

type sparkServiceStreamJobLogsServer struct {
	grpc.ServerStream
}

func (x *sparkServiceStreamJobLogsServer) Send(m *StreamJobLogsResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _SparkService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SparkServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/spark.v1.SparkService/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SparkServiceServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// SparkService_ServiceDesc is the grpc.ServiceDesc for SparkService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var SparkService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "spark.v1.SparkService",
	HandlerType: (*SparkServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RunJob",
			Handler:    _SparkService_RunJob_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _SparkService_Ping_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamJobLogs",
			Handler:       _SparkService_StreamJobLogs_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "spark/v1/spark_service.proto",
}
