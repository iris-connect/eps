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
	"encoding/base64"
	"fmt"
	"github.com/iris-gateway/eps"
	"github.com/iris-gateway/eps/jsonrpc"
	"github.com/iris-gateway/eps/tls"
	"github.com/kiprotect/go-helpers/forms"
	"net"
	"regexp"
	"time"
)

type PublicServer struct {
	settings             *PublicServerSettings
	jsonrpcServer        *jsonrpc.JSONRPCServer
	jsonrpcClient        *jsonrpc.Client
	tlsListener          net.Listener
	internalListener     net.Listener
	tlsConnections       map[string]net.Conn
	announcedConnections []*AnnouncedConnection
	tlsHellos            map[string][]byte
}

type AnnouncedConnection struct {
	Name    string         `json:"name"`
	Pattern *regexp.Regexp `json:"pattern"`
}

type IsValidRegexp struct{}

func (i IsValidRegexp) Validate(value interface{}, values map[string]interface{}) (interface{}, error) {
	// we assume IsString{} was called before...
	if regexp, err := regexp.Compile(value.(string)); err != nil {
		return nil, err
	} else {
		return regexp, nil
	}
}

var AnnounceConnectionForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "name",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name: "pattern",
			Validators: []forms.Validator{
				forms.IsString{},
				IsValidRegexp{},
			},
		},
	},
}

type AnnounceConnectionParams struct {
	Name    string `json:"name"`
	Pattern string `json:"pattern"`
}

func (c *PublicServer) announceConnection(context *jsonrpc.Context, params *AnnounceConnectionParams) *jsonrpc.Response {
	return context.InternalError()
}

func MakePublicServer(settings *PublicServerSettings) (*PublicServer, error) {
	server := &PublicServer{
		settings:             settings,
		jsonrpcClient:        jsonrpc.MakeClient(settings.JSONRPCClient),
		tlsConnections:       make(map[string]net.Conn),
		tlsHellos:            make(map[string][]byte),
		announcedConnections: make([]*AnnouncedConnection, 0),
	}

	methods := map[string]*jsonrpc.Method{
		"announceConnection": {
			Form:    &AnnounceConnectionForm,
			Handler: server.announceConnection,
		},
	}

	handler, err := jsonrpc.MethodsHandler(methods)

	if err != nil {
		return nil, err
	}

	jsonrpcServer, err := jsonrpc.MakeJSONRPCServer(settings.JSONRPCServer, handler)

	if err != nil {
		return nil, err
	}

	server.jsonrpcServer = jsonrpcServer

	return server, nil
}

func (s *PublicServer) jsonrpcHandler(context *jsonrpc.Context) *jsonrpc.Response {
	return nil
}

func (s *PublicServer) handleInternalConnection(internalConnection net.Conn) {
	// we give the client 1 second to announce itself
	internalConnection.SetReadDeadline(time.Now().Add(1 * time.Second))
	// we expect a secret token to be transmitted over the connection
	tokenBuf := make([]byte, 32)

	reqLen, err := internalConnection.Read(tokenBuf)

	if err != nil {
		eps.Log.Error(err)
	}

	if reqLen != 32 {
		eps.Log.Error("cannot read token")
	}

	tokenStr := base64.StdEncoding.EncodeToString(tokenBuf)

	tlsConnection, ok := s.tlsConnections[tokenStr]

	if !ok {
		return
	}

	tlsHello, ok := s.tlsHellos[tokenStr]

	if !ok {
		return
	}

	eps.Log.Info("Forwarding connection!")

	if n, err := internalConnection.Write(tlsHello); err != nil {
		eps.Log.Error(err)
		return
	} else if n != len(tlsHello) {
		eps.Log.Error("Can't forward TLS HelloClient")
		return
	} else {
		eps.Log.Info("Forwarded client hello")
	}

	pipe := func(left, right net.Conn) {
		buf := make([]byte, 1024)
		for {
			n, err := left.Read(buf)
			if err != nil {
				eps.Log.Error(err)
				return
			}
			if m, err := right.Write(buf[:n]); err != nil {
				eps.Log.Error(err)
				return
			} else if m != n {
				eps.Log.Errorf("cannot write all data")
			}
		}
	}

	go pipe(internalConnection, tlsConnection)
	go pipe(tlsConnection, internalConnection)

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

	close := func() {
		if err := conn.Close(); err != nil {
			eps.Log.Error(err)
		}
	}

	if err != nil {
		eps.Log.Error(err)
		close()
		return
	}

	if serverNameList := clientHello.ServerNameList(); serverNameList == nil {
		// no server name given, we close the connection
		close()
		return
	} else if hostName := serverNameList.HostName(); hostName == "" {
		close()
		return
	} else {

		id := fmt.Sprintf("%d", 1)

		randomBytes, err := jsonrpc.RandomBytes(32)

		if err != nil {
			close()
			return
		}

		randomStr := base64.StdEncoding.EncodeToString(randomBytes)

		// we store the connection details for later use
		s.tlsConnections[randomStr] = conn
		s.tlsHellos[randomStr] = buf[:reqLen]

		// we tell the internal proxy about an incoming connection
		request := jsonrpc.MakeRequest("private-proxy-1.incomingConnection", id, map[string]interface{}{
			"hostname": hostName,
			"token":    randomStr,
			"endpoint": s.settings.InternalBindAddress,
		})

		if result, err := s.jsonrpcClient.Call(request); err != nil {
			eps.Log.Error(err)
		} else {
			eps.Log.Info(result)
		}
	}
}

func (s *PublicServer) listenForTlsConnections() {
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
	go s.listenForTlsConnections()
	s.internalListener, err = net.Listen("tcp", s.settings.InternalBindAddress)
	if err != nil {
		return err
	}
	go s.listenForInternalConnections()

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
