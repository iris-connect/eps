package grpc

import (
	"github.com/iris-gateway/eps"
	"github.com/iris-gateway/eps/protobuf"
	"io"
)

type EPSServer struct {
	protobuf.UnimplementedEPSServer
}

func (s *EPSServer) MessageExchange(stream protobuf.EPS_MessageExchangeServer) error {
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		eps.Log.Info("Received message!")
	}
}

func MakeEPSServer() *EPSServer {
	return &EPSServer{}
}
