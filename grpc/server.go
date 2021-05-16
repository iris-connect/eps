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

package grpc

import (
	"context"
	"fmt"
	"github.com/iris-gateway/eps"
	"github.com/iris-gateway/eps/protobuf"
	"github.com/iris-gateway/eps/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/structpb"
	"net"
	"sync"
)

type ConnectedClient struct {
	CallServer protobuf.EPS_ServerCallServer
	Stop       chan bool
	Info       *eps.ClientInfo
	mutex      sync.Mutex
}

type Server struct {
	protobuf.UnimplementedEPSServer
	server           *grpc.Server
	settings         *GRPCServerSettings
	connectedClients map[string]*ConnectedClient
	directory        eps.Directory
	mutex            sync.Mutex
	handler          Handler
}

func (s *Server) Start() error {

	lis, err := net.Listen("tcp", s.settings.BindAddress)

	if err != nil {
		return err
	}

	go func() {
		s.server.Serve(lis)
	}()

	return nil

}

func (s *Server) Stop() error {
	return nil
}

func MakeServer(settings *GRPCServerSettings, handler Handler, directory eps.Directory) (*Server, error) {
	var opts []grpc.ServerOption
	if tlsConfig, err := tls.TLSServerConfig(settings.TLS); err != nil {
		return nil, err
	} else {
		opts = append(opts, grpc.Creds(VerifyCredentials{ClientInfo: &eps.ClientInfo{}, TransportCredentials: credentials.NewTLS(tlsConfig)}))
	}

	server := &Server{
		handler:          handler,
		directory:        directory,
		connectedClients: make(map[string]*ConnectedClient),
		server:           grpc.NewServer(opts...),
		settings:         settings,
	}

	protobuf.RegisterEPSServer(server.server, server)

	return server, nil
}

func (c *ConnectedClient) DeliverRequest(request *eps.Request) (*eps.Response, error) {

	// we need to ensure only one goroutine calls this method at once
	c.mutex.Lock()
	defer c.mutex.Unlock()

	paramsStruct, err := structpb.NewStruct(request.Params)

	if err != nil {
		return nil, err
	}

	pbRequest := &protobuf.Request{
		Params: paramsStruct,
		Method: request.Method,
		Id:     request.ID,
	}

	if err := c.CallServer.Send(pbRequest); err != nil {
		return nil, err
	}

	if pbResponse, err := c.CallServer.Recv(); err != nil {
		return nil, err
	} else {

		var responseError *eps.Error

		if pbResponse.Error != nil {
			responseError = &eps.Error{
				Code:    int(pbResponse.Error.Code),
				Data:    pbResponse.Error.Data.AsMap(),
				Message: pbResponse.Error.Message,
			}
		}

		response := &eps.Response{
			ID:     &pbResponse.Id,
			Result: pbResponse.Result.AsMap(),
			Error:  responseError,
		}

		return response, nil
	}

}

type Handler interface {
	HandleRequest(*eps.Request, *eps.ClientInfo) (*eps.Response, error)
}

func (s *Server) DeliverRequest(request *eps.Request) (*eps.Response, error) {

	address, err := eps.GetAddress(request.ID)

	if err != nil {
		return nil, err
	}

	client := s.getClient(address.Operator)

	if client == nil {
		return nil, fmt.Errorf("client disconnected")
	}

	return client.DeliverRequest(request)
}

func (s *Server) CanDeliverTo(address *eps.Address) bool {
	for _, connectedClient := range s.connectedClients {
		if connectedClient.Info.Name == address.Operator {
			return true
		}
	}
	return false
}

func (s *Server) Call(context context.Context, pbRequest *protobuf.Request) (*protobuf.Response, error) {

	clientInfo := &eps.ClientInfo{}

	peer, ok := peer.FromContext(context)
	if ok {
		tlsInfo := peer.AuthInfo.(credentials.TLSInfo)
		clientInfo.Name = tlsInfo.State.VerifiedChains[0][0].Subject.CommonName
	} else {
		return nil, fmt.Errorf("cannot determine client name")
	}

	request := &eps.Request{
		ID:     pbRequest.Id,
		Params: pbRequest.Params.AsMap(),
		Method: pbRequest.Method,
	}

	if response, err := s.handler.HandleRequest(request, clientInfo); err != nil {
		return nil, err
	} else {

		pbResponse := &protobuf.Response{
			Id: pbRequest.Id,
		}
		if response.Result != nil {
			resultStruct, err := structpb.NewStruct(response.Result)
			if err != nil {
				return nil, err
			}
			pbResponse.Result = resultStruct
		}
		if response.Error != nil {
			pbResponse.Error = &protobuf.Error{
				Code:    int32(response.Error.Code),
				Message: response.Error.Message,
			}

			if response.Error.Data != nil {
				errorStruct, err := structpb.NewStruct(response.Error.Data)
				if err != nil {
					return nil, err
				}
				pbResponse.Error.Data = errorStruct
			}

		}

		return pbResponse, nil

	}

}

func (s *Server) getClient(name string) *ConnectedClient {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	client, _ := s.connectedClients[name]
	return client
}

func (s *Server) setClient(client *ConnectedClient) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.connectedClients[client.Info.Name] = client
}

func (s *Server) ServerCall(server protobuf.EPS_ServerCallServer) error {

	// this is a bidirectional message stream

	clientInfo := &eps.ClientInfo{}

	peer, ok := peer.FromContext(server.Context())
	if ok {
		tlsInfo := peer.AuthInfo.(credentials.TLSInfo)
		clientInfo.Name = tlsInfo.State.VerifiedChains[0][0].Subject.CommonName
	} else {
		return fmt.Errorf("cannot determine client name")
	}

	client := s.getClient(clientInfo.Name)

	if client == nil {
		client = &ConnectedClient{
			Info: clientInfo,
			Stop: make(chan bool),
		}
		s.setClient(client)
	}

	client.CallServer = server

	// we wait for the client to stop...
	<-client.Stop

	return nil

}
