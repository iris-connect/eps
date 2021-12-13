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
	"github.com/iris-connect/eps"
	epsForms "github.com/iris-connect/eps/forms"
	"github.com/iris-connect/eps/jsonrpc"
	"github.com/kiprotect/go-helpers/forms"
	"regexp"
	"sync"
)

type Server struct {
	settings      *Settings
	jsonrpcServer *jsonrpc.JSONRPCServer
	directory     *RecordDirectory
	mutex         sync.Mutex
}

var SubmitChangeRecordsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "records",
			Validators: []forms.Validator{
				forms.IsList{
					Validators: []forms.Validator{
						forms.IsStringMap{
							Form: &epsForms.SignedChangeRecordForm,
						},
					},
				},
			},
		},
	},
}

type SubmitChangeRecordsParams struct {
	Records []*eps.SignedChangeRecord `json:"records"`
}

func (c *Server) submitChangeRecords(context *jsonrpc.Context, params *SubmitChangeRecordsParams) *jsonrpc.Response {
	if err := c.directory.Append(params.Records); err != nil {
		eps.Log.Error(err)
		return context.InternalError()
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
	After string `json:"after"`
}

var GetRecordsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "after",
			Validators: []forms.Validator{
				forms.IsOptional{Default: ""},
				forms.IsString{},
				forms.MatchesRegex{
					Regexp: regexp.MustCompile(`^([a-f0-9]{64}|)$`),
				},
			},
		},
	},
}

func (c *Server) getRecords(context *jsonrpc.Context, params *GetRecordsParams) *jsonrpc.Response {
	eps.Log.Infof("Getting records after '%s'", params.After)
	if records, err := c.directory.Records(params.After); err != nil {
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

	server.directory, err = MakeRecordDirectory(settings.Directory, settings.Definitions)

	if err != nil {
		return nil, err
	}

	methods := map[string]*jsonrpc.Method{
		"submitRecords": {
			Form:    &SubmitChangeRecordsForm,
			Handler: server.submitChangeRecords,
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
