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

package sd

import (
	"github.com/iris-gateway/eps/jsonrpc"
	"github.com/kiprotect/go-helpers/forms"
	"sync"
)

type Server struct {
	settings      *Settings
	jsonrpcServer *jsonrpc.JSONRPCServer
	directory     *RecordDirectory
	mutex         sync.Mutex
}

var SubmitChangeRecordForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "name",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
	},
}

type SubmitChangeRecordParams struct {
	Name    string `json:"name"`
	Pattern string `json:"pattern"`
}

func (c *Server) submitChangeRecord(context *jsonrpc.Context, params *SubmitChangeRecordParams) *jsonrpc.Response {
	return context.InternalError()
}

var GetTipForm = forms.Form{
	Fields: []forms.Field{},
}

type GetTipParams struct {
}

func (c *Server) getTip(context *jsonrpc.Context, params *GetTipParams) *jsonrpc.Response {
	return context.InternalError()
}

func MakeServer(settings *Settings) (*Server, error) {
	server := &Server{
		settings: settings,
	}

	var err error

	server.directory, err = MakeRecordDirectory(settings.Directory)

	if err != nil {
		return nil, err
	}

	methods := map[string]*jsonrpc.Method{
		"submitChangeRecord": {
			Form:    &SubmitChangeRecordForm,
			Handler: server.submitChangeRecord,
		},
		"getTip": {
			Form:    &GetTipForm,
			Handler: server.getTip,
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

func (s *Server) Start() error {
	return s.jsonrpcServer.Start()
}

func (s *Server) Stop() error {
	return s.jsonrpcServer.Stop()
}
