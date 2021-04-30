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
	server   *grpc.Server
	settings *eps.GRPCServerSettings
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

func MakeServer(settings *eps.GRPCServerSettings) (*Server, error) {
	var opts []grpc.ServerOption
	if tlsConfig, err := tls.TLSServerConfig(settings.TLS); err != nil {
		return nil, err
	} else {
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}
	grpcServer := grpc.NewServer(opts...)
	protobuf.RegisterEPSServer(grpcServer, MakeEPSServer())
	return &Server{
		server:   grpcServer,
		settings: settings,
	}, nil
}
