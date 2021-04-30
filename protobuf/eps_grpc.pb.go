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
	MessageExchange(ctx context.Context, opts ...grpc.CallOption) (EPS_MessageExchangeClient, error)
}

type ePSClient struct {
	cc grpc.ClientConnInterface
}

func NewEPSClient(cc grpc.ClientConnInterface) EPSClient {
	return &ePSClient{cc}
}

func (c *ePSClient) MessageExchange(ctx context.Context, opts ...grpc.CallOption) (EPS_MessageExchangeClient, error) {
	stream, err := c.cc.NewStream(ctx, &EPS_ServiceDesc.Streams[0], "/EPS/MessageExchange", opts...)
	if err != nil {
		return nil, err
	}
	x := &ePSMessageExchangeClient{stream}
	return x, nil
}

type EPS_MessageExchangeClient interface {
	Send(*Message) error
	Recv() (*Message, error)
	grpc.ClientStream
}

type ePSMessageExchangeClient struct {
	grpc.ClientStream
}

func (x *ePSMessageExchangeClient) Send(m *Message) error {
	return x.ClientStream.SendMsg(m)
}

func (x *ePSMessageExchangeClient) Recv() (*Message, error) {
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
	MessageExchange(EPS_MessageExchangeServer) error
	mustEmbedUnimplementedEPSServer()
}

// UnimplementedEPSServer must be embedded to have forward compatible implementations.
type UnimplementedEPSServer struct {
}

func (UnimplementedEPSServer) MessageExchange(EPS_MessageExchangeServer) error {
	return status.Errorf(codes.Unimplemented, "method MessageExchange not implemented")
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

func _EPS_MessageExchange_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(EPSServer).MessageExchange(&ePSMessageExchangeServer{stream})
}

type EPS_MessageExchangeServer interface {
	Send(*Message) error
	Recv() (*Message, error)
	grpc.ServerStream
}

type ePSMessageExchangeServer struct {
	grpc.ServerStream
}

func (x *ePSMessageExchangeServer) Send(m *Message) error {
	return x.ServerStream.SendMsg(m)
}

func (x *ePSMessageExchangeServer) Recv() (*Message, error) {
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
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "MessageExchange",
			Handler:       _EPS_MessageExchange_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "protobuf/eps.proto",
}
