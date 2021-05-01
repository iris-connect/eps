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
	Call(ctx context.Context, in *Message, opts ...grpc.CallOption) (*Message, error)
	Stream(ctx context.Context, opts ...grpc.CallOption) (EPS_StreamClient, error)
}

type ePSClient struct {
	cc grpc.ClientConnInterface
}

func NewEPSClient(cc grpc.ClientConnInterface) EPSClient {
	return &ePSClient{cc}
}

func (c *ePSClient) Call(ctx context.Context, in *Message, opts ...grpc.CallOption) (*Message, error) {
	out := new(Message)
	err := c.cc.Invoke(ctx, "/EPS/Call", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ePSClient) Stream(ctx context.Context, opts ...grpc.CallOption) (EPS_StreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &EPS_ServiceDesc.Streams[0], "/EPS/Stream", opts...)
	if err != nil {
		return nil, err
	}
	x := &ePSStreamClient{stream}
	return x, nil
}

type EPS_StreamClient interface {
	Send(*Message) error
	Recv() (*Message, error)
	grpc.ClientStream
}

type ePSStreamClient struct {
	grpc.ClientStream
}

func (x *ePSStreamClient) Send(m *Message) error {
	return x.ClientStream.SendMsg(m)
}

func (x *ePSStreamClient) Recv() (*Message, error) {
	m := new(Message)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// EPSServer is the server API for EPS service.
// All implementations must embed UnimplementedEPSServer
// for forward compatibility
type EPSServer interface {
	Call(context.Context, *Message) (*Message, error)
	Stream(EPS_StreamServer) error
	mustEmbedUnimplementedEPSServer()
}

// UnimplementedEPSServer must be embedded to have forward compatible implementations.
type UnimplementedEPSServer struct {
}

func (UnimplementedEPSServer) Call(context.Context, *Message) (*Message, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Call not implemented")
}
func (UnimplementedEPSServer) Stream(EPS_StreamServer) error {
	return status.Errorf(codes.Unimplemented, "method Stream not implemented")
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
	in := new(Message)
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
		return srv.(EPSServer).Call(ctx, req.(*Message))
	}
	return interceptor(ctx, in, info, handler)
}

func _EPS_Stream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(EPSServer).Stream(&ePSStreamServer{stream})
}

type EPS_StreamServer interface {
	Send(*Message) error
	Recv() (*Message, error)
	grpc.ServerStream
}

type ePSStreamServer struct {
	grpc.ServerStream
}

func (x *ePSStreamServer) Send(m *Message) error {
	return x.ServerStream.SendMsg(m)
}

func (x *ePSStreamServer) Recv() (*Message, error) {
	m := new(Message)
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
			StreamName:    "Stream",
			Handler:       _EPS_Stream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "protobuf/eps.proto",
}
