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
	"github.com/iris-gateway/eps"
	"github.com/iris-gateway/eps/protobuf"
	"github.com/iris-gateway/eps/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
)

type Server struct {
	server    *grpc.Server
	epsServer *EPSServer
	settings  *GRPCServerSettings
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

func (s *Server) CanDeliverTo(address *eps.Address) bool {
	return s.epsServer.CanDeliverTo(address)
}

func (s *Server) DeliverRequest(request *eps.Request) (*eps.Response, error) {
	return s.epsServer.DeliverRequest(request)
}

func MakeServer(settings *GRPCServerSettings, handler Handler) (*Server, error) {
	var opts []grpc.ServerOption
	if tlsConfig, err := tls.TLSServerConfig(settings.TLS); err != nil {
		return nil, err
	} else {
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}

	epsServer := MakeEPSServer(handler)

	grpcServer := grpc.NewServer(opts...)
	protobuf.RegisterEPSServer(grpcServer, epsServer)
	return &Server{
		epsServer: epsServer,
		server:    grpcServer,
		settings:  settings,
	}, nil
}
