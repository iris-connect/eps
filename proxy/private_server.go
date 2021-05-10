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
The private proy server creates outgoing TCP connections to the public proxy
server and forwards them to an internal endpoint.
*/

package proxy

import (
	"encoding/base64"
	"fmt"
	"github.com/iris-gateway/eps"
	"github.com/iris-gateway/eps/jsonrpc"
	"net"
)

type PrivateServer struct {
	settings      *PrivateServerSettings
	jsonrpcServer *jsonrpc.JSONRPCServer
	jsonrpcClient *jsonrpc.Client
	l             net.Listener
}

type ProxyConnection struct {
	proxyEndpoint    string
	internalEndpoint string
	token            []byte
}

func MakeProxyConnection(proxyEndpoint, internalEndpoint string, token []byte) *ProxyConnection {
	return &ProxyConnection{
		proxyEndpoint:    proxyEndpoint,
		internalEndpoint: internalEndpoint,
		token:            token,
	}
}

func (p *ProxyConnection) Run() error {

	proxyConnection, err := net.Dial("tcp", p.proxyEndpoint)

	if err != nil {
		return err
	}

	if n, err := proxyConnection.Write(p.token); err != nil {
		return err
	} else if n != len(p.token) {
		return fmt.Errorf("could not write token")
	}

	internalConnection, err := net.Dial("tcp", p.internalEndpoint)

	if err != nil {
		return err
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

	go pipe(internalConnection, proxyConnection)
	go pipe(proxyConnection, internalConnection)

	return nil
}

func MakePrivateServer(settings *PrivateServerSettings) (*PrivateServer, error) {

	server := &PrivateServer{
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

func (s *PrivateServer) jsonrpcHandler(context *jsonrpc.Context) *jsonrpc.Response {
	params := context.Request.Params
	tokenStr := params["token"].(string)
	proxyEndpoint := params["endpoint"].(string)
	token, err := base64.StdEncoding.DecodeString(tokenStr)
	if err != nil {
		eps.Log.Error(err)
		return context.InternalError()
	}

	connection := MakeProxyConnection(proxyEndpoint, s.settings.InternalEndpoint, token)

	if err := connection.Run(); err != nil {
		eps.Log.Error(err)
		return context.InternalError()
	}

	return context.Result(map[string]interface{}{"message": "ok"})
}

func (s *PrivateServer) Start() error {
	eps.Log.Debug("Starting JSON-RPC server")
	return s.jsonrpcServer.Start()
}

func (s *PrivateServer) Stop() error {
	return s.jsonrpcServer.Stop()
}
