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
	// client sends a request to the server and receives a response
	Call(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
	// client sends a response to the server and receives an acknowledgment
	Respond(ctx context.Context, in *Response, opts ...grpc.CallOption) (*Acknowledgment, error)
	// server sends requests to the client and receives responses
	ServerCall(ctx context.Context, opts ...grpc.CallOption) (EPS_ServerCallClient, error)
	// server sends responses to the client and receives acknowledgments
	ServerRespond(ctx context.Context, opts ...grpc.CallOption) (EPS_ServerRespondClient, error)
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

func (c *ePSClient) Respond(ctx context.Context, in *Response, opts ...grpc.CallOption) (*Acknowledgment, error) {
	out := new(Acknowledgment)
	err := c.cc.Invoke(ctx, "/EPS/Respond", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ePSClient) ServerCall(ctx context.Context, opts ...grpc.CallOption) (EPS_ServerCallClient, error) {
	stream, err := c.cc.NewStream(ctx, &EPS_ServiceDesc.Streams[0], "/EPS/ServerCall", opts...)
	if err != nil {
		return nil, err
	}
	x := &ePSServerCallClient{stream}
	return x, nil
}

type EPS_ServerCallClient interface {
	Send(*Response) error
	Recv() (*Request, error)
	grpc.ClientStream
}

type ePSServerCallClient struct {
	grpc.ClientStream
}

func (x *ePSServerCallClient) Send(m *Response) error {
	return x.ClientStream.SendMsg(m)
}

func (x *ePSServerCallClient) Recv() (*Request, error) {
	m := new(Request)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *ePSClient) ServerRespond(ctx context.Context, opts ...grpc.CallOption) (EPS_ServerRespondClient, error) {
	stream, err := c.cc.NewStream(ctx, &EPS_ServiceDesc.Streams[1], "/EPS/ServerRespond", opts...)
	if err != nil {
		return nil, err
	}
	x := &ePSServerRespondClient{stream}
	return x, nil
}

type EPS_ServerRespondClient interface {
	Send(*Acknowledgment) error
	Recv() (*Response, error)
	grpc.ClientStream
}

type ePSServerRespondClient struct {
	grpc.ClientStream
}

func (x *ePSServerRespondClient) Send(m *Acknowledgment) error {
	return x.ClientStream.SendMsg(m)
}

func (x *ePSServerRespondClient) Recv() (*Response, error) {
	m := new(Response)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// EPSServer is the server API for EPS service.
// All implementations must embed UnimplementedEPSServer
// for forward compatibility
type EPSServer interface {
	// client sends a request to the server and receives a response
	Call(context.Context, *Request) (*Response, error)
	// client sends a response to the server and receives an acknowledgment
	Respond(context.Context, *Response) (*Acknowledgment, error)
	// server sends requests to the client and receives responses
	ServerCall(EPS_ServerCallServer) error
	// server sends responses to the client and receives acknowledgments
	ServerRespond(EPS_ServerRespondServer) error
	mustEmbedUnimplementedEPSServer()
}

// UnimplementedEPSServer must be embedded to have forward compatible implementations.
type UnimplementedEPSServer struct {
}

func (UnimplementedEPSServer) Call(context.Context, *Request) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Call not implemented")
}
func (UnimplementedEPSServer) Respond(context.Context, *Response) (*Acknowledgment, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Respond not implemented")
}
func (UnimplementedEPSServer) ServerCall(EPS_ServerCallServer) error {
	return status.Errorf(codes.Unimplemented, "method ServerCall not implemented")
}
func (UnimplementedEPSServer) ServerRespond(EPS_ServerRespondServer) error {
	return status.Errorf(codes.Unimplemented, "method ServerRespond not implemented")
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

func _EPS_Respond_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Response)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EPSServer).Respond(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/EPS/Respond",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EPSServer).Respond(ctx, req.(*Response))
	}
	return interceptor(ctx, in, info, handler)
}

func _EPS_ServerCall_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(EPSServer).ServerCall(&ePSServerCallServer{stream})
}

type EPS_ServerCallServer interface {
	Send(*Request) error
	Recv() (*Response, error)
	grpc.ServerStream
}

type ePSServerCallServer struct {
	grpc.ServerStream
}

func (x *ePSServerCallServer) Send(m *Request) error {
	return x.ServerStream.SendMsg(m)
}

func (x *ePSServerCallServer) Recv() (*Response, error) {
	m := new(Response)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _EPS_ServerRespond_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(EPSServer).ServerRespond(&ePSServerRespondServer{stream})
}

type EPS_ServerRespondServer interface {
	Send(*Response) error
	Recv() (*Acknowledgment, error)
	grpc.ServerStream
}

type ePSServerRespondServer struct {
	grpc.ServerStream
}

func (x *ePSServerRespondServer) Send(m *Response) error {
	return x.ServerStream.SendMsg(m)
}

func (x *ePSServerRespondServer) Recv() (*Acknowledgment, error) {
	m := new(Acknowledgment)
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
		{
			MethodName: "Respond",
			Handler:    _EPS_Respond_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ServerCall",
			Handler:       _EPS_ServerCall_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "ServerRespond",
			Handler:       _EPS_ServerRespond_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "protobuf/eps.proto",
}
