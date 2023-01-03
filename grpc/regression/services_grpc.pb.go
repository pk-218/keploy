// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.12
// source: grpc/regression/services.proto

package regression

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

// RegressionServiceClient is the client API for RegressionService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RegressionServiceClient interface {
	End(ctx context.Context, in *EndRequest, opts ...grpc.CallOption) (*EndResponse, error)
	Start(ctx context.Context, in *StartRequest, opts ...grpc.CallOption) (*StartResponse, error)
	GetTC(ctx context.Context, in *GetTCRequest, opts ...grpc.CallOption) (*TestCase, error)
	GetTCS(ctx context.Context, in *GetTCSRequest, opts ...grpc.CallOption) (*GetTCSResponse, error)
	PostTC(ctx context.Context, in *TestCaseReq, opts ...grpc.CallOption) (*PostTCResponse, error)
	DeNoise(ctx context.Context, in *TestReq, opts ...grpc.CallOption) (*DeNoiseResponse, error)
	Test(ctx context.Context, in *TestReq, opts ...grpc.CallOption) (*TestResponse, error)
	PutMock(ctx context.Context, in *PutMockReq, opts ...grpc.CallOption) (*PutMockResp, error)
	GetMocks(ctx context.Context, in *GetMockReq, opts ...grpc.CallOption) (*GetMockResp, error)
	StartMocking(ctx context.Context, in *StartMockReq, opts ...grpc.CallOption) (*StartMockResp, error)
}

type regressionServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewRegressionServiceClient(cc grpc.ClientConnInterface) RegressionServiceClient {
	return &regressionServiceClient{cc}
}

func (c *regressionServiceClient) End(ctx context.Context, in *EndRequest, opts ...grpc.CallOption) (*EndResponse, error) {
	out := new(EndResponse)
	err := c.cc.Invoke(ctx, "/services.RegressionService/End", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *regressionServiceClient) Start(ctx context.Context, in *StartRequest, opts ...grpc.CallOption) (*StartResponse, error) {
	out := new(StartResponse)
	err := c.cc.Invoke(ctx, "/services.RegressionService/Start", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *regressionServiceClient) GetTC(ctx context.Context, in *GetTCRequest, opts ...grpc.CallOption) (*TestCase, error) {
	out := new(TestCase)
	err := c.cc.Invoke(ctx, "/services.RegressionService/GetTC", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *regressionServiceClient) GetTCS(ctx context.Context, in *GetTCSRequest, opts ...grpc.CallOption) (*GetTCSResponse, error) {
	out := new(GetTCSResponse)
	err := c.cc.Invoke(ctx, "/services.RegressionService/GetTCS", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *regressionServiceClient) PostTC(ctx context.Context, in *TestCaseReq, opts ...grpc.CallOption) (*PostTCResponse, error) {
	out := new(PostTCResponse)
	err := c.cc.Invoke(ctx, "/services.RegressionService/PostTC", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *regressionServiceClient) DeNoise(ctx context.Context, in *TestReq, opts ...grpc.CallOption) (*DeNoiseResponse, error) {
	out := new(DeNoiseResponse)
	err := c.cc.Invoke(ctx, "/services.RegressionService/DeNoise", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *regressionServiceClient) Test(ctx context.Context, in *TestReq, opts ...grpc.CallOption) (*TestResponse, error) {
	out := new(TestResponse)
	err := c.cc.Invoke(ctx, "/services.RegressionService/Test", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *regressionServiceClient) PutMock(ctx context.Context, in *PutMockReq, opts ...grpc.CallOption) (*PutMockResp, error) {
	out := new(PutMockResp)
	err := c.cc.Invoke(ctx, "/services.RegressionService/PutMock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *regressionServiceClient) GetMocks(ctx context.Context, in *GetMockReq, opts ...grpc.CallOption) (*GetMockResp, error) {
	out := new(GetMockResp)
	err := c.cc.Invoke(ctx, "/services.RegressionService/GetMocks", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *regressionServiceClient) StartMocking(ctx context.Context, in *StartMockReq, opts ...grpc.CallOption) (*StartMockResp, error) {
	out := new(StartMockResp)
	err := c.cc.Invoke(ctx, "/services.RegressionService/StartMocking", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RegressionServiceServer is the server API for RegressionService service.
// All implementations must embed UnimplementedRegressionServiceServer
// for forward compatibility
type RegressionServiceServer interface {
	End(context.Context, *EndRequest) (*EndResponse, error)
	Start(context.Context, *StartRequest) (*StartResponse, error)
	GetTC(context.Context, *GetTCRequest) (*TestCase, error)
	GetTCS(context.Context, *GetTCSRequest) (*GetTCSResponse, error)
	PostTC(context.Context, *TestCaseReq) (*PostTCResponse, error)
	DeNoise(context.Context, *TestReq) (*DeNoiseResponse, error)
	Test(context.Context, *TestReq) (*TestResponse, error)
	PutMock(context.Context, *PutMockReq) (*PutMockResp, error)
	GetMocks(context.Context, *GetMockReq) (*GetMockResp, error)
	StartMocking(context.Context, *StartMockReq) (*StartMockResp, error)
	mustEmbedUnimplementedRegressionServiceServer()
}

// UnimplementedRegressionServiceServer must be embedded to have forward compatible implementations.
type UnimplementedRegressionServiceServer struct {
}

func (UnimplementedRegressionServiceServer) End(context.Context, *EndRequest) (*EndResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method End not implemented")
}
func (UnimplementedRegressionServiceServer) Start(context.Context, *StartRequest) (*StartResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Start not implemented")
}
func (UnimplementedRegressionServiceServer) GetTC(context.Context, *GetTCRequest) (*TestCase, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTC not implemented")
}
func (UnimplementedRegressionServiceServer) GetTCS(context.Context, *GetTCSRequest) (*GetTCSResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTCS not implemented")
}
func (UnimplementedRegressionServiceServer) PostTC(context.Context, *TestCaseReq) (*PostTCResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostTC not implemented")
}
func (UnimplementedRegressionServiceServer) DeNoise(context.Context, *TestReq) (*DeNoiseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeNoise not implemented")
}
func (UnimplementedRegressionServiceServer) Test(context.Context, *TestReq) (*TestResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Test not implemented")
}
func (UnimplementedRegressionServiceServer) PutMock(context.Context, *PutMockReq) (*PutMockResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PutMock not implemented")
}
func (UnimplementedRegressionServiceServer) GetMocks(context.Context, *GetMockReq) (*GetMockResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMocks not implemented")
}
func (UnimplementedRegressionServiceServer) StartMocking(context.Context, *StartMockReq) (*StartMockResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartMocking not implemented")
}
func (UnimplementedRegressionServiceServer) mustEmbedUnimplementedRegressionServiceServer() {}

// UnsafeRegressionServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RegressionServiceServer will
// result in compilation errors.
type UnsafeRegressionServiceServer interface {
	mustEmbedUnimplementedRegressionServiceServer()
}

func RegisterRegressionServiceServer(s grpc.ServiceRegistrar, srv RegressionServiceServer) {
	s.RegisterService(&RegressionService_ServiceDesc, srv)
}

func _RegressionService_End_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EndRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RegressionServiceServer).End(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/services.RegressionService/End",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RegressionServiceServer).End(ctx, req.(*EndRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RegressionService_Start_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StartRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RegressionServiceServer).Start(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/services.RegressionService/Start",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RegressionServiceServer).Start(ctx, req.(*StartRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RegressionService_GetTC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTCRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RegressionServiceServer).GetTC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/services.RegressionService/GetTC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RegressionServiceServer).GetTC(ctx, req.(*GetTCRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RegressionService_GetTCS_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTCSRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RegressionServiceServer).GetTCS(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/services.RegressionService/GetTCS",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RegressionServiceServer).GetTCS(ctx, req.(*GetTCSRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RegressionService_PostTC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TestCaseReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RegressionServiceServer).PostTC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/services.RegressionService/PostTC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RegressionServiceServer).PostTC(ctx, req.(*TestCaseReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _RegressionService_DeNoise_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TestReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RegressionServiceServer).DeNoise(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/services.RegressionService/DeNoise",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RegressionServiceServer).DeNoise(ctx, req.(*TestReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _RegressionService_Test_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TestReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RegressionServiceServer).Test(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/services.RegressionService/Test",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RegressionServiceServer).Test(ctx, req.(*TestReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _RegressionService_PutMock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutMockReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RegressionServiceServer).PutMock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/services.RegressionService/PutMock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RegressionServiceServer).PutMock(ctx, req.(*PutMockReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _RegressionService_GetMocks_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMockReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RegressionServiceServer).GetMocks(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/services.RegressionService/GetMocks",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RegressionServiceServer).GetMocks(ctx, req.(*GetMockReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _RegressionService_StartMocking_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StartMockReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RegressionServiceServer).StartMocking(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/services.RegressionService/StartMocking",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RegressionServiceServer).StartMocking(ctx, req.(*StartMockReq))
	}
	return interceptor(ctx, in, info, handler)
}

// RegressionService_ServiceDesc is the grpc.ServiceDesc for RegressionService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RegressionService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "services.RegressionService",
	HandlerType: (*RegressionServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "End",
			Handler:    _RegressionService_End_Handler,
		},
		{
			MethodName: "Start",
			Handler:    _RegressionService_Start_Handler,
		},
		{
			MethodName: "GetTC",
			Handler:    _RegressionService_GetTC_Handler,
		},
		{
			MethodName: "GetTCS",
			Handler:    _RegressionService_GetTCS_Handler,
		},
		{
			MethodName: "PostTC",
			Handler:    _RegressionService_PostTC_Handler,
		},
		{
			MethodName: "DeNoise",
			Handler:    _RegressionService_DeNoise_Handler,
		},
		{
			MethodName: "Test",
			Handler:    _RegressionService_Test_Handler,
		},
		{
			MethodName: "PutMock",
			Handler:    _RegressionService_PutMock_Handler,
		},
		{
			MethodName: "GetMocks",
			Handler:    _RegressionService_GetMocks_Handler,
		},
		{
			MethodName: "StartMocking",
			Handler:    _RegressionService_StartMocking_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "grpc/regression/services.proto",
}
