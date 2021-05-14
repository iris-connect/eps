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
	"github.com/iris-gateway/eps"
	epsForms "github.com/iris-gateway/eps/forms"
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
			Name: "record",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &epsForms.SignedChangeRecordForm,
				},
			},
		},
	},
}

type SubmitChangeRecordParams struct {
	Record *eps.SignedChangeRecord `json:"record"`
}

func (c *Server) submitChangeRecord(context *jsonrpc.Context, params *SubmitChangeRecordParams) *jsonrpc.Response {
	eps.Log.Info(params.Record.Record.CreatedAt)
	if err := c.directory.Append(params.Record); err != nil {
		eps.Log.Error(err)
		return context.Error(400, "something went wrong", nil)
	} else {
		return context.Acknowledge()
	}
}

var GetTipForm = forms.Form{
	Fields: []forms.Field{},
}

type GetTipParams struct {
}

func (c *Server) getTip(context *jsonrpc.Context, params *GetTipParams) *jsonrpc.Response {
	if record, err := c.directory.Tip(); err != nil {
		eps.Log.Error(err)
		return context.InternalError()
	} else {
		return context.Result(record)
	}
}

type GetEntriesParams struct {
}

var GetEntriesForm = forms.Form{
	Fields: []forms.Field{},
}

func (c *Server) getEntries(context *jsonrpc.Context, params *GetEntriesParams) *jsonrpc.Response {
	if entries, err := c.directory.AllEntries(); err != nil {
		eps.Log.Error(err)
		return context.InternalError()
	} else {
		return context.Result(entries)
	}
}

type GetEntryParams struct {
	Name string `json:"name"`
}

var GetEntryForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "name",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
	},
}

func (c *Server) getEntry(context *jsonrpc.Context, params *GetEntryParams) *jsonrpc.Response {
	if entry, err := c.directory.Entry(params.Name); err != nil {
		eps.Log.Error(err)
		return context.InternalError()
	} else if entry == nil {
		return context.NotFound()
	} else {
		return context.Result(entry)
	}
}

type GetRecordsParams struct {
	Since int64 `json:"since"`
}

var GetRecordsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "since",
			Validators: []forms.Validator{
				forms.IsOptional{Default: 0},
				forms.IsInteger{
					HasMin: true,
					Min:    0,
				},
			},
		},
	},
}

func (c *Server) getRecords(context *jsonrpc.Context, params *GetRecordsParams) *jsonrpc.Response {
	if records, err := c.directory.Records(params.Since); err != nil {
		return context.InternalError()
	} else {
		return context.Result(records)
	}
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
		"getRecords": {
			Form:    &GetRecordsForm,
			Handler: server.getRecords,
		},
		"getEntries": {
			Form:    &GetEntriesForm,
			Handler: server.getEntries,
		},
		"getEntry": {
			Form:    &GetEntryForm,
			Handler: server.getEntry,
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
