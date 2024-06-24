// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.25.1
// source: contract/contract.proto

package contract

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

// DFSClient is the client API for DFS service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DFSClient interface {
	StoreFile(ctx context.Context, opts ...grpc.CallOption) (DFS_StoreFileClient, error)
	FetchFile(ctx context.Context, in *MetaData, opts ...grpc.CallOption) (DFS_FetchFileClient, error)
	ListFiles(ctx context.Context, in *MetaData, opts ...grpc.CallOption) (*ListResponse, error)
	GetFileStat(ctx context.Context, in *MetaData, opts ...grpc.CallOption) (*MetaData, error)
	LockFile(ctx context.Context, in *MetaData, opts ...grpc.CallOption) (*MetaData, error)
	CallbackList(ctx context.Context, in *MetaData, opts ...grpc.CallOption) (*ListResponse, error)
	DeleteFile(ctx context.Context, in *MetaData, opts ...grpc.CallOption) (*MetaData, error)
}

type dFSClient struct {
	cc grpc.ClientConnInterface
}

func NewDFSClient(cc grpc.ClientConnInterface) DFSClient {
	return &dFSClient{cc}
}

func (c *dFSClient) StoreFile(ctx context.Context, opts ...grpc.CallOption) (DFS_StoreFileClient, error) {
	stream, err := c.cc.NewStream(ctx, &DFS_ServiceDesc.Streams[0], "/contract.DFS/StoreFile", opts...)
	if err != nil {
		return nil, err
	}
	x := &dFSStoreFileClient{stream}
	return x, nil
}

type DFS_StoreFileClient interface {
	Send(*StoreRequest) error
	CloseAndRecv() (*MetaData, error)
	grpc.ClientStream
}

type dFSStoreFileClient struct {
	grpc.ClientStream
}

func (x *dFSStoreFileClient) Send(m *StoreRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *dFSStoreFileClient) CloseAndRecv() (*MetaData, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(MetaData)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *dFSClient) FetchFile(ctx context.Context, in *MetaData, opts ...grpc.CallOption) (DFS_FetchFileClient, error) {
	stream, err := c.cc.NewStream(ctx, &DFS_ServiceDesc.Streams[1], "/contract.DFS/FetchFile", opts...)
	if err != nil {
		return nil, err
	}
	x := &dFSFetchFileClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type DFS_FetchFileClient interface {
	Recv() (*FetchResponse, error)
	grpc.ClientStream
}

type dFSFetchFileClient struct {
	grpc.ClientStream
}

func (x *dFSFetchFileClient) Recv() (*FetchResponse, error) {
	m := new(FetchResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *dFSClient) ListFiles(ctx context.Context, in *MetaData, opts ...grpc.CallOption) (*ListResponse, error) {
	out := new(ListResponse)
	err := c.cc.Invoke(ctx, "/contract.DFS/ListFiles", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dFSClient) GetFileStat(ctx context.Context, in *MetaData, opts ...grpc.CallOption) (*MetaData, error) {
	out := new(MetaData)
	err := c.cc.Invoke(ctx, "/contract.DFS/GetFileStat", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dFSClient) LockFile(ctx context.Context, in *MetaData, opts ...grpc.CallOption) (*MetaData, error) {
	out := new(MetaData)
	err := c.cc.Invoke(ctx, "/contract.DFS/LockFile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dFSClient) CallbackList(ctx context.Context, in *MetaData, opts ...grpc.CallOption) (*ListResponse, error) {
	out := new(ListResponse)
	err := c.cc.Invoke(ctx, "/contract.DFS/CallbackList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dFSClient) DeleteFile(ctx context.Context, in *MetaData, opts ...grpc.CallOption) (*MetaData, error) {
	out := new(MetaData)
	err := c.cc.Invoke(ctx, "/contract.DFS/DeleteFile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DFSServer is the server API for DFS service.
// All implementations must embed UnimplementedDFSServer
// for forward compatibility
type DFSServer interface {
	StoreFile(DFS_StoreFileServer) error
	FetchFile(*MetaData, DFS_FetchFileServer) error
	ListFiles(context.Context, *MetaData) (*ListResponse, error)
	GetFileStat(context.Context, *MetaData) (*MetaData, error)
	LockFile(context.Context, *MetaData) (*MetaData, error)
	CallbackList(context.Context, *MetaData) (*ListResponse, error)
	DeleteFile(context.Context, *MetaData) (*MetaData, error)
	mustEmbedUnimplementedDFSServer()
}

// UnimplementedDFSServer must be embedded to have forward compatible implementations.
type UnimplementedDFSServer struct {
}

func (UnimplementedDFSServer) StoreFile(DFS_StoreFileServer) error {
	return status.Errorf(codes.Unimplemented, "method StoreFile not implemented")
}
func (UnimplementedDFSServer) FetchFile(*MetaData, DFS_FetchFileServer) error {
	return status.Errorf(codes.Unimplemented, "method FetchFile not implemented")
}
func (UnimplementedDFSServer) ListFiles(context.Context, *MetaData) (*ListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListFiles not implemented")
}
func (UnimplementedDFSServer) GetFileStat(context.Context, *MetaData) (*MetaData, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFileStat not implemented")
}
func (UnimplementedDFSServer) LockFile(context.Context, *MetaData) (*MetaData, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LockFile not implemented")
}
func (UnimplementedDFSServer) CallbackList(context.Context, *MetaData) (*ListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CallbackList not implemented")
}
func (UnimplementedDFSServer) DeleteFile(context.Context, *MetaData) (*MetaData, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteFile not implemented")
}
func (UnimplementedDFSServer) mustEmbedUnimplementedDFSServer() {}

// UnsafeDFSServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DFSServer will
// result in compilation errors.
type UnsafeDFSServer interface {
	mustEmbedUnimplementedDFSServer()
}

func RegisterDFSServer(s grpc.ServiceRegistrar, srv DFSServer) {
	s.RegisterService(&DFS_ServiceDesc, srv)
}

func _DFS_StoreFile_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(DFSServer).StoreFile(&dFSStoreFileServer{stream})
}

type DFS_StoreFileServer interface {
	SendAndClose(*MetaData) error
	Recv() (*StoreRequest, error)
	grpc.ServerStream
}

type dFSStoreFileServer struct {
	grpc.ServerStream
}

func (x *dFSStoreFileServer) SendAndClose(m *MetaData) error {
	return x.ServerStream.SendMsg(m)
}

func (x *dFSStoreFileServer) Recv() (*StoreRequest, error) {
	m := new(StoreRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _DFS_FetchFile_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(MetaData)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(DFSServer).FetchFile(m, &dFSFetchFileServer{stream})
}

type DFS_FetchFileServer interface {
	Send(*FetchResponse) error
	grpc.ServerStream
}

type dFSFetchFileServer struct {
	grpc.ServerStream
}

func (x *dFSFetchFileServer) Send(m *FetchResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _DFS_ListFiles_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MetaData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DFSServer).ListFiles(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/contract.DFS/ListFiles",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DFSServer).ListFiles(ctx, req.(*MetaData))
	}
	return interceptor(ctx, in, info, handler)
}

func _DFS_GetFileStat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MetaData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DFSServer).GetFileStat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/contract.DFS/GetFileStat",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DFSServer).GetFileStat(ctx, req.(*MetaData))
	}
	return interceptor(ctx, in, info, handler)
}

func _DFS_LockFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MetaData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DFSServer).LockFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/contract.DFS/LockFile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DFSServer).LockFile(ctx, req.(*MetaData))
	}
	return interceptor(ctx, in, info, handler)
}

func _DFS_CallbackList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MetaData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DFSServer).CallbackList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/contract.DFS/CallbackList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DFSServer).CallbackList(ctx, req.(*MetaData))
	}
	return interceptor(ctx, in, info, handler)
}

func _DFS_DeleteFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MetaData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DFSServer).DeleteFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/contract.DFS/DeleteFile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DFSServer).DeleteFile(ctx, req.(*MetaData))
	}
	return interceptor(ctx, in, info, handler)
}

// DFS_ServiceDesc is the grpc.ServiceDesc for DFS service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DFS_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "contract.DFS",
	HandlerType: (*DFSServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListFiles",
			Handler:    _DFS_ListFiles_Handler,
		},
		{
			MethodName: "GetFileStat",
			Handler:    _DFS_GetFileStat_Handler,
		},
		{
			MethodName: "LockFile",
			Handler:    _DFS_LockFile_Handler,
		},
		{
			MethodName: "CallbackList",
			Handler:    _DFS_CallbackList_Handler,
		},
		{
			MethodName: "DeleteFile",
			Handler:    _DFS_DeleteFile_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StoreFile",
			Handler:       _DFS_StoreFile_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "FetchFile",
			Handler:       _DFS_FetchFile_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "contract/contract.proto",
}
