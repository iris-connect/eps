// IRIS Endpoint-Server (EPS)
// Copyright (C) 2021-2021 The IRIS Endpoint-Server Authors (see AUTHORS.md)
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package protobuf

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

// EPSClient is the client API for EPS service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EPSClient interface {
	// synchronous method calls
	Call(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
	// async RPCs
	AsyncUpstream(ctx context.Context, opts ...grpc.CallOption) (EPS_AsyncUpstreamClient, error)
	// async reverse RPCs
	AsyncDownstream(ctx context.Context, opts ...grpc.CallOption) (EPS_AsyncDownstreamClient, error)
}

type ePSClient struct {
	cc grpc.ClientConnInterface
}

func NewEPSClient(cc grpc.ClientConnInterface) EPSClient {
	return &ePSClient{cc}
}

func (c *ePSClient) Call(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/EPS/Call", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ePSClient) AsyncUpstream(ctx context.Context, opts ...grpc.CallOption) (EPS_AsyncUpstreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &EPS_ServiceDesc.Streams[0], "/EPS/AsyncUpstream", opts...)
	if err != nil {
		return nil, err
	}
	x := &ePSAsyncUpstreamClient{stream}
	return x, nil
}

type EPS_AsyncUpstreamClient interface {
	Send(*Request) error
	Recv() (*Response, error)
	grpc.ClientStream
}

type ePSAsyncUpstreamClient struct {
	grpc.ClientStream
}

func (x *ePSAsyncUpstreamClient) Send(m *Request) error {
	return x.ClientStream.SendMsg(m)
}

func (x *ePSAsyncUpstreamClient) Recv() (*Response, error) {
	m := new(Response)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *ePSClient) AsyncDownstream(ctx context.Context, opts ...grpc.CallOption) (EPS_AsyncDownstreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &EPS_ServiceDesc.Streams[1], "/EPS/AsyncDownstream", opts...)
	if err != nil {
		return nil, err
	}
	x := &ePSAsyncDownstreamClient{stream}
	return x, nil
}

type EPS_AsyncDownstreamClient interface {
	Send(*Response) error
	Recv() (*Request, error)
	grpc.ClientStream
}

type ePSAsyncDownstreamClient struct {
	grpc.ClientStream
}

func (x *ePSAsyncDownstreamClient) Send(m *Response) error {
	return x.ClientStream.SendMsg(m)
}

func (x *ePSAsyncDownstreamClient) Recv() (*Request, error) {
	m := new(Request)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// EPSServer is the server API for EPS service.
// All implementations must embed UnimplementedEPSServer
// for forward compatibility
type EPSServer interface {
	// synchronous method calls
	Call(context.Context, *Request) (*Response, error)
	// async RPCs
	AsyncUpstream(EPS_AsyncUpstreamServer) error
	// async reverse RPCs
	AsyncDownstream(EPS_AsyncDownstreamServer) error
	mustEmbedUnimplementedEPSServer()
}

// UnimplementedEPSServer must be embedded to have forward compatible implementations.
type UnimplementedEPSServer struct {
}

func (UnimplementedEPSServer) Call(context.Context, *Request) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Call not implemented")
}
func (UnimplementedEPSServer) AsyncUpstream(EPS_AsyncUpstreamServer) error {
	return status.Errorf(codes.Unimplemented, "method AsyncUpstream not implemented")
}
func (UnimplementedEPSServer) AsyncDownstream(EPS_AsyncDownstreamServer) error {
	return status.Errorf(codes.Unimplemented, "method AsyncDownstream not implemented")
}
func (UnimplementedEPSServer) mustEmbedUnimplementedEPSServer() {}

// UnsafeEPSServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EPSServer will
// result in compilation errors.
type UnsafeEPSServer interface {
	mustEmbedUnimplementedEPSServer()
}

func RegisterEPSServer(s grpc.ServiceRegistrar, srv EPSServer) {
	s.RegisterService(&EPS_ServiceDesc, srv)
}

func _EPS_Call_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EPSServer).Call(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/EPS/Call",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EPSServer).Call(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _EPS_AsyncUpstream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(EPSServer).AsyncUpstream(&ePSAsyncUpstreamServer{stream})
}

type EPS_AsyncUpstreamServer interface {
	Send(*Response) error
	Recv() (*Request, error)
	grpc.ServerStream
}

type ePSAsyncUpstreamServer struct {
	grpc.ServerStream
}

func (x *ePSAsyncUpstreamServer) Send(m *Response) error {
	return x.ServerStream.SendMsg(m)
}

func (x *ePSAsyncUpstreamServer) Recv() (*Request, error) {
	m := new(Request)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _EPS_AsyncDownstream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(EPSServer).AsyncDownstream(&ePSAsyncDownstreamServer{stream})
}

type EPS_AsyncDownstreamServer interface {
	Send(*Request) error
	Recv() (*Response, error)
	grpc.ServerStream
}

type ePSAsyncDownstreamServer struct {
	grpc.ServerStream
}

func (x *ePSAsyncDownstreamServer) Send(m *Request) error {
	return x.ServerStream.SendMsg(m)
}

func (x *ePSAsyncDownstreamServer) Recv() (*Response, error) {
	m := new(Response)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// EPS_ServiceDesc is the grpc.ServiceDesc for EPS service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var EPS_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "EPS",
	HandlerType: (*EPSServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Call",
			Handler:    _EPS_Call_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "AsyncUpstream",
			Handler:       _EPS_AsyncUpstream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "AsyncDownstream",
			Handler:       _EPS_AsyncDownstream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "protobuf/eps.proto",
}
