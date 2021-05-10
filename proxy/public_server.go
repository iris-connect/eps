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

/*
The public proxy accepts incoming TLS connections (using a TCP connection),
parses the `HelloClient` packet and forwards the connection to the internal
proxy via a separate TCP channel.
*/

package proxy

import (
	"fmt"
	"github.com/iris-gateway/eps"
	"github.com/iris-gateway/eps/jsonrpc"
	"github.com/iris-gateway/eps/tls"
	"net"
	"time"
)

type PublicServer struct {
	settings         *PublicServerSettings
	jsonrpcServer    *jsonrpc.JSONRPCServer
	jsonrpcClient    *jsonrpc.Client
	tlsListener      net.Listener
	internalListener net.Listener
}

func MakePublicServer(settings *PublicServerSettings) (*PublicServer, error) {
	server := &PublicServer{
		settings:      settings,
		jsonrpcClient: jsonrpc.MakeClient(settings.JSONRPCClient),
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

func (s *PublicServer) handleInternalConnection(conn net.Conn) {
	// we give the client 1 second to announce itself
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	// we expect a secret token to be transmitted over the connection
	buf := make([]byte, 32)

	reqLen, err := conn.Read(buf)

	if err != nil {
		eps.Log.Error(err)
	}

}

func (s *PublicServer) handleTlsConnection(conn net.Conn) {
	// we give the client 1 second to announce itself
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))

	// 2 kB is more than enough for a TLS ClientHello packet
	buf := make([]byte, 2048)

	reqLen, err := conn.Read(buf)

	if err != nil {
		eps.Log.Error(err)
	}

	clientHello, err := tls.ParseClientHello(buf[:reqLen])

	if err != nil {
		eps.Log.Error(err)
		if err := conn.Close(); err != nil {
			eps.Log.Error(err)
		}
		return
	}

	if serverNameList := clientHello.ServerNameList(); serverNameList == nil {
		// no server name given, we close the connection
		if err := conn.Close(); err != nil {
			eps.Log.Error(err)
		}
	} else if hostName := serverNameList.HostName(); hostName == "" {
		// no host name given, we close the connection
		if err := conn.Close(); err != nil {
			eps.Log.Error(err)
		}
	} else {

		id := fmt.Sprintf("%d", 1)

		request := jsonrpc.MakeRequest("private-proxy-1.incomingConnection", id, map[string]interface{}{
			"hostname": hostName,
		})

		if result, err := s.jsonrpcClient.Call(request); err != nil {
			eps.Log.Error(err)
		} else {
			eps.Log.Info(result)
		}
	}
}

func (s *PublicServer) listenForTls() {
	for {
		if s.tlsListener == nil {
			// was closed
			break
		}
		conn, err := s.tlsListener.Accept()
		if err != nil {
			if err == net.ErrClosed {
				break
			}
			eps.Log.Error("another error")
			eps.Log.Error(err)
			continue
		}
		go s.handleTlsConnection(conn)
	}
}

func (s *PublicServer) listenForInternalConnections() {
	for {
		if s.internalListener == nil {
			// was closed
			break
		}
		conn, err := s.internalListener.Accept()
		if err != nil {
			if err == net.ErrClosed {
				break
			}
			eps.Log.Error(err)
			continue
		}
		go s.handleInternalConnection(conn)
	}

}

func (s *PublicServer) Start() error {
	var err error
	s.tlsListener, err = net.Listen("tcp", s.settings.TLSBindAddress)
	if err != nil {
		return err
	}
	go s.listenForTls()

	if err := s.jsonrpcServer.Start(); err != nil {
		return err
	}

	return nil
}

func (s *PublicServer) Stop() error {
	if s.tlsListener != nil {
		if err := s.tlsListener.Close(); err != nil {
			eps.Log.Error(err)
		}
		s.tlsListener = nil
	}
	return s.jsonrpcServer.Stop()
}
