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

// +build examples

package main

import (
	"github.com/iris-connect/eps"
	"github.com/iris-connect/eps/jsonrpc"
	"github.com/kiprotect/go-helpers/forms"
	"os"
	"os/signal"
	"syscall"
)

var locationsDB = make(map[string]*Location)

func handler(context *jsonrpc.Context) *jsonrpc.Response {
	switch context.Request.Method {
	case "add":
		if params, err := AddLocationForm.Validate(context.Request.Params); err != nil {
			return context.InvalidParams(err)
		} else {
			location := &Location{}
			if err := AddLocationForm.Coerce(location, params); err != nil {
				eps.Log.Error(err)
				return context.InternalError()
			}
			locationsDB[params["id"].(string)] = location
			return context.Result("ok")
		}
	case "lookup":
		if params, err := LookupLocationForm.Validate(context.Request.Params); err != nil {
			return context.InvalidParams(err)
		} else {
			name := params["name"].(string)
			for _, location := range locationsDB {
				if location.Name == name {
					return context.Result(location)
				}
			}
			return context.Error(2, "not found", "foo")
		}
	}
	return context.MethodNotFound()
}

var LookupLocationForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "name",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
	},
}

var AddLocationForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "name",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name: "id",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
	},
}

type Location struct {
	Name string `json:"name"`
	ID string `json:"id"`
}

func main() {

	settings := &jsonrpc.JSONRPCServerSettings{
		BindAddress: "localhost:6666",
	}

	if server, err := jsonrpc.MakeJSONRPCServer(settings, handler); err != nil {
		eps.Log.Fatal(err)
	} else {
		metrics.OpenPrometheusEndpoint()
		server.Start()

		// we wait for CTRL-C / Interrupt
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

		eps.Log.Info("Waiting for CTRL-C...")

		<-sigchan

		eps.Log.Info("Stopping server...")

		server.Stop()

	}
}
