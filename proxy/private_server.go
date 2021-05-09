package proxy

import (
	"github.com/iris-gateway/eps"
	"github.com/iris-gateway/eps/jsonrpc"
	"net"
)

type PrivateServer struct {
	settings      *PrivateServerSettings
	jsonrpcServer *jsonrpc.JSONRPCServer
	l             net.Listener
}

func MakePrivateServer(settings *PrivateServerSettings) (*PrivateServer, error) {

	server := &PrivateServer{
		settings: settings,
	}

	jsonrpcServer, err := jsonrpc.MakeJSONRPCServer(settings.JSONRPCServer, server.jsonrpcHandler)

	if err != nil {
		return nil, err
	}

	server.jsonrpcServer = jsonrpcServer

	return server, nil

}

func (s *PrivateServer) jsonrpcHandler(context *jsonrpc.Context) *jsonrpc.Response {
	return nil
}

func (s *PrivateServer) handle(conn net.Conn) {
	buf := make([]byte, 1024*100)

	reqLen, err := conn.Read(buf)

	if err != nil {
		eps.Log.Error(err)
	}

	eps.Log.Infof("%d: %s", reqLen, string(buf))
}

func (s *PrivateServer) listen() {
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

func (s *PrivateServer) Start() error {
	var err error
	s.l, err = net.Listen("tcp", s.settings.BindAddress)
	if err != nil {
		return err
	}
	go s.listen()
	return nil
}

func (s *PrivateServer) Stop() error {
	if s.l != nil {
		s.l.Close()
		s.l = nil
	}
	return nil
}
