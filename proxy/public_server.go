package proxy

import (
	"encoding/hex"
	"github.com/iris-gateway/eps"
	"github.com/iris-gateway/eps/jsonrpc"
	"net"
)

type PublicServer struct {
	settings      *PublicServerSettings
	jsonrpcServer *jsonrpc.JSONRPCServer
	l             net.Listener
}

func MakePublicServer(settings *PublicServerSettings) (*PublicServer, error) {
	server := &PublicServer{
		settings: settings,
	}

	jsonrpcServer, err := jsonrpc.MakeJSONRPCServer(settings.JSONRPCServer, server.jsonrpcHandler)

	if err != nil {
		return nil, err
	}

	server.jsonrpcServer = jsonrpcServer

	return server, nil
}

func (s *PublicServer) jsonrpcHandler(context *jsonrpc.Context) *jsonrpc.Response {
	return nil
}

func (s *PublicServer) handle(conn net.Conn) {
	buf := make([]byte, 1024)

	reqLen, err := conn.Read(buf)

	if err != nil {
		eps.Log.Error(err)
	}

	eps.Log.Infof(hex.EncodeToString(buf[:reqLen]))

	eps.Log.Infof("%d", reqLen)
}

func (s *PublicServer) listen() {
	eps.Log.Info("Listeing...")
	for {
		conn, err := s.l.Accept()
		if err != nil {
			eps.Log.Error(err)
		}
		eps.Log.Info("Accepted request.")
		s.handle(conn)
	}
}

func (s *PublicServer) Start() error {
	var err error
	s.l, err = net.Listen("tcp", s.settings.BindAddress)
	if err != nil {
		return err
	}
	go s.listen()
	return nil
}

func (s *PublicServer) Stop() error {
	if s.l != nil {
		s.l.Close()
		s.l = nil
	}
	return nil
}
