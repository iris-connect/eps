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
	"github.com/iris-connect/eps"
	"github.com/iris-connect/eps/helpers"
	epsNet "github.com/iris-connect/eps/net"
	"github.com/iris-connect/eps/protobuf"
	"github.com/iris-connect/eps/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/structpb"
	"net"
	"sync"
	"time"
)

type ConnectedClient struct {
	CallServer protobuf.EPS_ServerCallServer
	Stop       chan bool
	directory  eps.Directory
	Info       *eps.ClientInfo
	mutex      sync.Mutex
}

type Server struct {
	protobuf.UnimplementedEPSServer
	server           *grpc.Server
	settings         *GRPCServerSettings
	connectedClients []*ConnectedClient
	directory        eps.Directory
	mutex            sync.Mutex
	handler          Handler
}

func (s *Server) Start() error {

	lis, err := net.Listen("tcp", s.settings.BindAddress)

	if err != nil {
		return fmt.Errorf("error binding to address '%s': %w", s.settings.BindAddress, err)
	}

	if s.settings.TCPRateLimits != nil {
		lis = epsNet.MakeRateLimitedListener(lis, s.settings.TCPRateLimits)
	}

	go func() {
		s.server.Serve(lis)
	}()

	return nil

}

func (s *Server) Stop() error {
	return nil
}

// currently we allow messages up to 4MB in size
var MaxMessageSize = 1024 * 1024 * 4

func MakeServer(settings *GRPCServerSettings, handler Handler, directory eps.Directory) (*Server, error) {
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(MaxMessageSize),
		grpc.MaxSendMsgSize(MaxMessageSize),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{MinTime: 15 * time.Second, PermitWithoutStream: true}),
		grpc.KeepaliveParams(keepalive.ServerParameters{MaxConnectionIdle: 60 * time.Second, MaxConnectionAge: 24 * time.Hour, MaxConnectionAgeGrace: 1 * time.Minute, Time: 1 * time.Minute, Timeout: 30 * time.Second}),
	}
	if tlsConfig, err := tls.TLSServerConfig(settings.TLS); err != nil {
		return nil, fmt.Errorf("error retrieving TLS server config: %w", err)
	} else {
		opts = append(opts, grpc.Creds(VerifyCredentials{directory: directory, TransportCredentials: credentials.NewTLS(tlsConfig)}))
	}

	server := &Server{
		handler:          handler,
		directory:        directory,
		connectedClients: []*ConnectedClient{},
		server:           grpc.NewServer(opts...),
		settings:         settings,
	}

	protobuf.RegisterEPSServer(server.server, server)

	return server, nil
}

func (c *ConnectedClient) DeliverRequest(request *eps.Request) (*eps.Response, error) {

	eps.Log.Debugf("Trying to deliver request to connected client '%s'...", c.Info.Name)

	// we need to ensure only one goroutine calls this method at once
	c.mutex.Lock()
	defer c.mutex.Unlock()

	paramsStruct, err := structpb.NewStruct(request.Params)

	if err != nil {
		c.Stop <- true
		return nil, fmt.Errorf("error serializing params for gRPC: %w", err)
	}

	pbRequest := &protobuf.Request{
		ClientName: c.directory.Name(),
		Params:     paramsStruct,
		Method:     request.Method,
		Id:         request.ID,
	}

	if err := c.CallServer.Send(pbRequest); err != nil {
		eps.Log.Errorf("Cannot deliver request: %v", err)
		c.Stop <- true
		return nil, fmt.Errorf("error sending gRPC request: %w", err)
	}

	if pbResponse, err := c.CallServer.Recv(); err != nil {
		eps.Log.Errorf("Cannot receive response: %v", err)
		// we close the connection
		c.Stop <- true
		return nil, fmt.Errorf("error receiving gRPC response: %w", err)
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
		return nil, fmt.Errorf("error parsing address: %w", err)
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

	peer, ok := peer.FromContext(context)

	if !ok {
		return nil, fmt.Errorf("cannot get peer")
	}

	clientInfoAuthInfo, ok := peer.AuthInfo.(*ClientInfoAuthInfo)

	if !ok {
		return nil, fmt.Errorf("cannot determine client info")
	}

	request := &eps.Request{
		ID:     pbRequest.Id,
		Params: pbRequest.Params.AsMap(),
		Method: pbRequest.Method,
	}

	// we make sure the name that the client has given matches with one name
	// from the certificate...
	clientInfo := clientInfoAuthInfo.ClientInfos.ClientInfo(pbRequest.ClientName)

	if clientInfo == nil {
		return nil, fmt.Errorf("no matching client")
	}

	if response, err := s.handler.HandleRequest(request, clientInfo); err != nil {
		return nil, fmt.Errorf("error handling gRPC request: %w", err)
	} else {

		pbResponse := &protobuf.Response{
			Id: pbRequest.Id,
		}
		if response != nil {
			if response.Result != nil {
				stringMap, err := helpers.ToStringMap(response.Result)
				if err != nil {
					return nil, fmt.Errorf("error converting result to string map: %w", err)
				}
				resultStruct, err := structpb.NewStruct(stringMap)
				if err != nil {
					return nil, fmt.Errorf("error serializing response for gRPC: %w", err)
				}
				pbResponse.Result = resultStruct
			}
			if response.Error != nil {
				pbResponse.Error = &protobuf.Error{
					Code:    int32(response.Error.Code),
					Message: response.Error.Message,
				}

				if response.Error.Data != nil {
					stringMap, err := helpers.ToStringMap(response.Error.Data)
					if err != nil {
						return nil, fmt.Errorf("error converting error data to string map: %w", err)
					}
					errorStruct, err := structpb.NewStruct(stringMap)
					if err != nil {
						return nil, fmt.Errorf("error serializing error data for gRPC: %w", err)
					}
					pbResponse.Error.Data = errorStruct
				}

			}
		}
		return pbResponse, nil

	}

}

func (s *Server) getClient(name string) *ConnectedClient {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for _, client := range s.connectedClients {
		if client.Info.Name == name {
			return client
		}
	}
	return nil
}

func (s *Server) setClient(client *ConnectedClient) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	newClients := []*ConnectedClient{}
	for _, existingClient := range s.connectedClients {
		if existingClient == client {
			continue
		}
		newClients = append(newClients, existingClient)
	}
	newClients = append(newClients, client)
	s.connectedClients = newClients
}

func (s *Server) deleteClient(client *ConnectedClient) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	newClients := []*ConnectedClient{}
	for _, existingClient := range s.connectedClients {
		if existingClient == client {
			continue
		}
		newClients = append(newClients, existingClient)
	}
	s.connectedClients = newClients
}

type ClientAnnouncement struct {
	Name string `json:"name"`
}

func (s *Server) ServerCall(server protobuf.EPS_ServerCallServer) error {

	// this is a bidirectional message stream

	peer, ok := peer.FromContext(server.Context())

	if !ok {
		return fmt.Errorf("cannot get peer")
	}

	clientInfoAuthInfo, ok := peer.AuthInfo.(*ClientInfoAuthInfo)

	if !ok {
		return fmt.Errorf("cannot determine client info")
	}

	pbResponse, err := server.Recv()

	if err != nil {
		eps.Log.Error(err)
		return fmt.Errorf("can't receive handshake packet: %v", err)
	}

	data := pbResponse.Result.AsMap()
	clientAnnouncement := &ClientAnnouncement{}

	if params, err := AnnouncementForm.Validate(data); err != nil {
		return fmt.Errorf("invalid client announcement")
	} else if err := AnnouncementForm.Coerce(clientAnnouncement, params); err != nil {
		return err
	}

	name := clientAnnouncement.Name

	eps.Log.Debugf("Client announced itself as '%s'", name)

	if !clientInfoAuthInfo.ClientInfos.HasName(name) {
		return fmt.Errorf("invalid client name supplied")
	}

	client := s.getClient(name)

	if client == nil {
		client = &ConnectedClient{
			Info:       clientInfoAuthInfo.ClientInfos.ClientInfo(name),
			Stop:       make(chan bool),
			CallServer: server,
			directory:  s.directory,
		}
		s.setClient(client)
	}

	eps.Log.Debugf("Received incoming gRPC connection from client '%s' (primary name)", clientInfoAuthInfo.ClientInfos.PrimaryName())

	// we update the CallServer reference in the client (in case it has been updated)
	s.mutex.Lock()
	client.CallServer = server
	s.mutex.Unlock()

	// we wait for the client to stop...
	select {
	case <-client.Stop:
		break
	// the server is done (e.g. because the connection was closed)
	case <-server.Context().Done():
		break
	}

	s.deleteClient(client)

	return nil

}
