// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.6
// source: engine_service.proto

package kurtosis_engine_rpc_api_bindings

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// EngineServiceClient is the client API for EngineService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EngineServiceClient interface {
	// Endpoint for getting information about the engine, which is also what we use to verify that the engine has become available
	GetEngineInfo(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetEngineInfoResponse, error)
	// ==============================================================================================
	//                                   Enclave Management
	// ==============================================================================================
	// Creates a new Kurtosis Enclave
	CreateEnclave(ctx context.Context, in *CreateEnclaveArgs, opts ...grpc.CallOption) (*CreateEnclaveResponse, error)
	// Returns information about the existing enclaves
	GetEnclaves(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetEnclavesResponse, error)
	// Stops all containers in an enclave
	StopEnclave(ctx context.Context, in *StopEnclaveArgs, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// Destroys an enclave, removing all artifacts associated with it
	DestroyEnclave(ctx context.Context, in *DestroyEnclaveArgs, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// Gets rid of old enclaves
	Clean(ctx context.Context, in *CleanArgs, opts ...grpc.CallOption) (*CleanResponse, error)
	// Get service logs
	GetServiceLogs(ctx context.Context, in *GetServiceLogsArgs, opts ...grpc.CallOption) (EngineService_GetServiceLogsClient, error)
}

type engineServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewEngineServiceClient(cc grpc.ClientConnInterface) EngineServiceClient {
	return &engineServiceClient{cc}
}

func (c *engineServiceClient) GetEngineInfo(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetEngineInfoResponse, error) {
	out := new(GetEngineInfoResponse)
	err := c.cc.Invoke(ctx, "/engine_api.EngineService/GetEngineInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *engineServiceClient) CreateEnclave(ctx context.Context, in *CreateEnclaveArgs, opts ...grpc.CallOption) (*CreateEnclaveResponse, error) {
	out := new(CreateEnclaveResponse)
	err := c.cc.Invoke(ctx, "/engine_api.EngineService/CreateEnclave", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *engineServiceClient) GetEnclaves(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetEnclavesResponse, error) {
	out := new(GetEnclavesResponse)
	err := c.cc.Invoke(ctx, "/engine_api.EngineService/GetEnclaves", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *engineServiceClient) StopEnclave(ctx context.Context, in *StopEnclaveArgs, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/engine_api.EngineService/StopEnclave", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *engineServiceClient) DestroyEnclave(ctx context.Context, in *DestroyEnclaveArgs, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/engine_api.EngineService/DestroyEnclave", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *engineServiceClient) Clean(ctx context.Context, in *CleanArgs, opts ...grpc.CallOption) (*CleanResponse, error) {
	out := new(CleanResponse)
	err := c.cc.Invoke(ctx, "/engine_api.EngineService/Clean", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *engineServiceClient) GetServiceLogs(ctx context.Context, in *GetServiceLogsArgs, opts ...grpc.CallOption) (EngineService_GetServiceLogsClient, error) {
	stream, err := c.cc.NewStream(ctx, &EngineService_ServiceDesc.Streams[0], "/engine_api.EngineService/GetServiceLogs", opts...)
	if err != nil {
		return nil, err
	}
	x := &engineServiceGetServiceLogsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type EngineService_GetServiceLogsClient interface {
	Recv() (*GetServiceLogsResponse, error)
	grpc.ClientStream
}

type engineServiceGetServiceLogsClient struct {
	grpc.ClientStream
}

func (x *engineServiceGetServiceLogsClient) Recv() (*GetServiceLogsResponse, error) {
	m := new(GetServiceLogsResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// EngineServiceServer is the server API for EngineService service.
// for forward compatibility
type EngineServiceServer interface {
	// Endpoint for getting information about the engine, which is also what we use to verify that the engine has become available
	GetEngineInfo(context.Context, *emptypb.Empty) (*GetEngineInfoResponse, error)
	// ==============================================================================================
	//                                   Enclave Management
	// ==============================================================================================
	// Creates a new Kurtosis Enclave
	CreateEnclave(context.Context, *CreateEnclaveArgs) (*CreateEnclaveResponse, error)
	// Returns information about the existing enclaves
	GetEnclaves(context.Context, *emptypb.Empty) (*GetEnclavesResponse, error)
	// Stops all containers in an enclave
	StopEnclave(context.Context, *StopEnclaveArgs) (*emptypb.Empty, error)
	// Destroys an enclave, removing all artifacts associated with it
	DestroyEnclave(context.Context, *DestroyEnclaveArgs) (*emptypb.Empty, error)
	// Gets rid of old enclaves
	Clean(context.Context, *CleanArgs) (*CleanResponse, error)
	// Get service logs
	GetServiceLogs(*GetServiceLogsArgs, EngineService_GetServiceLogsServer) error
}

type UnimplementedEngineServiceServer struct {
}

func (UnimplementedEngineServiceServer) GetEngineInfo(context.Context, *emptypb.Empty) (*GetEngineInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetEngineInfo not implemented")
}
func (UnimplementedEngineServiceServer) CreateEnclave(context.Context, *CreateEnclaveArgs) (*CreateEnclaveResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateEnclave not implemented")
}
func (UnimplementedEngineServiceServer) GetEnclaves(context.Context, *emptypb.Empty) (*GetEnclavesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetEnclaves not implemented")
}
func (UnimplementedEngineServiceServer) StopEnclave(context.Context, *StopEnclaveArgs) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StopEnclave not implemented")
}
func (UnimplementedEngineServiceServer) DestroyEnclave(context.Context, *DestroyEnclaveArgs) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DestroyEnclave not implemented")
}
func (UnimplementedEngineServiceServer) Clean(context.Context, *CleanArgs) (*CleanResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Clean not implemented")
}
func (UnimplementedEngineServiceServer) GetServiceLogs(*GetServiceLogsArgs, EngineService_GetServiceLogsServer) error {
	return status.Errorf(codes.Unimplemented, "method GetServiceLogs not implemented")
}
func (UnimplementedEngineServiceServer) mustEmbedUnimplementedEngineServiceServer() {}

// UnsafeEngineServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EngineServiceServer will
// result in compilation errors.
type UnsafeEngineServiceServer interface {
	mustEmbedUnimplementedEngineServiceServer()
}

func RegisterEngineServiceServer(s grpc.ServiceRegistrar, srv EngineServiceServer) {
	s.RegisterService(&EngineService_ServiceDesc, srv)
}

func _EngineService_GetEngineInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EngineServiceServer).GetEngineInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/engine_api.EngineService/GetEngineInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EngineServiceServer).GetEngineInfo(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _EngineService_CreateEnclave_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateEnclaveArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EngineServiceServer).CreateEnclave(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/engine_api.EngineService/CreateEnclave",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EngineServiceServer).CreateEnclave(ctx, req.(*CreateEnclaveArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _EngineService_GetEnclaves_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EngineServiceServer).GetEnclaves(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/engine_api.EngineService/GetEnclaves",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EngineServiceServer).GetEnclaves(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _EngineService_StopEnclave_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopEnclaveArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EngineServiceServer).StopEnclave(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/engine_api.EngineService/StopEnclave",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EngineServiceServer).StopEnclave(ctx, req.(*StopEnclaveArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _EngineService_DestroyEnclave_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DestroyEnclaveArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EngineServiceServer).DestroyEnclave(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/engine_api.EngineService/DestroyEnclave",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EngineServiceServer).DestroyEnclave(ctx, req.(*DestroyEnclaveArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _EngineService_Clean_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CleanArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EngineServiceServer).Clean(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/engine_api.EngineService/Clean",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EngineServiceServer).Clean(ctx, req.(*CleanArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _EngineService_GetServiceLogs_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetServiceLogsArgs)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(EngineServiceServer).GetServiceLogs(m, &engineServiceGetServiceLogsServer{stream})
}

type EngineService_GetServiceLogsServer interface {
	Send(*GetServiceLogsResponse) error
	grpc.ServerStream
}

type engineServiceGetServiceLogsServer struct {
	grpc.ServerStream
}

func (x *engineServiceGetServiceLogsServer) Send(m *GetServiceLogsResponse) error {
	return x.ServerStream.SendMsg(m)
}

// EngineService_ServiceDesc is the grpc.ServiceDesc for EngineService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var EngineService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "engine_api.EngineService",
	HandlerType: (*EngineServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetEngineInfo",
			Handler:    _EngineService_GetEngineInfo_Handler,
		},
		{
			MethodName: "CreateEnclave",
			Handler:    _EngineService_CreateEnclave_Handler,
		},
		{
			MethodName: "GetEnclaves",
			Handler:    _EngineService_GetEnclaves_Handler,
		},
		{
			MethodName: "StopEnclave",
			Handler:    _EngineService_StopEnclave_Handler,
		},
		{
			MethodName: "DestroyEnclave",
			Handler:    _EngineService_DestroyEnclave_Handler,
		},
		{
			MethodName: "Clean",
			Handler:    _EngineService_Clean_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetServiceLogs",
			Handler:       _EngineService_GetServiceLogs_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "engine_service.proto",
}
