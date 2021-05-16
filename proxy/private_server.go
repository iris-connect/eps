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
	"fmt"
	"github.com/iris-gateway/eps"
	epsForms "github.com/iris-gateway/eps/forms"
	"github.com/iris-gateway/eps/jsonrpc"
	"github.com/kiprotect/go-helpers/forms"
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
		proxyConnection.Close()
		return err
	} else if n != len(p.token) {
		proxyConnection.Close()
		return fmt.Errorf("could not write token")
	}

	internalConnection, err := net.Dial("tcp", p.internalEndpoint)

	if err != nil {
		proxyConnection.Close()
		return err
	}

	close := func() {
		proxyConnection.Close()
		internalConnection.Close()
	}

	pipe := func(left, right net.Conn) {
		buf := make([]byte, 1024)
		for {
			n, err := left.Read(buf)
			if err != nil {
				eps.Log.Error(err)
				close()
				return
			}
			if m, err := right.Write(buf[:n]); err != nil {
				eps.Log.Error(err)
				close()
				return
			} else if m != n {
				eps.Log.Errorf("cannot write all data")
				close()
				return
			}
		}
	}

	go pipe(internalConnection, proxyConnection)
	go pipe(proxyConnection, internalConnection)

	return nil
}

var IncomingConnectionForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "token",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MinLength: 32,
					MaxLength: 32,
				},
			},
		},
		{
			Name: "endpoint",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name: "_client",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &epsForms.ClientInfoForm,
				},
			},
		},
	},
}

type IncomingConnectionParams struct {
	Endpoint string          `json:"endpoint"`
	Token    []byte          `json:"token"`
	Client   *eps.ClientInfo `json:"_client"`
}

func (c *PrivateServer) incomingConnection(context *jsonrpc.Context, params *IncomingConnectionParams) *jsonrpc.Response {

	eps.Log.Info(params.Client)
	connection := MakeProxyConnection(params.Endpoint, c.settings.InternalEndpoint, params.Token)

	go func() {
		if err := connection.Run(); err != nil {
			eps.Log.Error(err)
		}
	}()

	return context.Result(map[string]interface{}{"message": "ok"})
}

func MakePrivateServer(settings *PrivateServerSettings) (*PrivateServer, error) {

	server := &PrivateServer{
		settings:      settings,
		jsonrpcClient: jsonrpc.MakeClient(settings.JSONRPCClient),
	}

	methods := map[string]*jsonrpc.Method{
		"incomingConnection": {
			Form:    &IncomingConnectionForm,
			Handler: server.incomingConnection,
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

func (s *PrivateServer) announceConnections() error {
	return nil
	/*
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
	*/
}

func (s *PrivateServer) Start() error {
	return s.jsonrpcServer.Start()
}

func (s *PrivateServer) Stop() error {
	return s.jsonrpcServer.Stop()
}
