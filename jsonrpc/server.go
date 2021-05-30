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

package jsonrpc

import (
	"fmt"
	"github.com/iris-connect/eps/http"
)

type Handler func(*Context) *Response

type JSONRPCServer struct {
	settings *JSONRPCServerSettings
	server   *http.HTTPServer
	handler  Handler
}

func JSONRPC(handler Handler) http.Handler {
	return func(c *http.Context) {
		// the request data has been validated by the 'ExtractJSONRequest' handler
		request := c.Get("request").(*Request)

		context := &Context{
			Request:     request,
			HTTPContext: c,
		}

		response := handler(context)

		if response == nil {
			response = context.Nil()
		}

		// people will forget this so we add it here in that case
		if response.JSONRPC == "" {
			response.JSONRPC = "2.0"
		}

		code := 200

		// if there was an error we return a 400 status instead of 200
		if response.Error != nil {
			code = 400
		}

		c.JSON(code, response)

	}
}

func NotFound(c *http.Context) {
	c.JSON(404, map[string]interface{}{"message": "not found"})
}

func MakeJSONRPCServer(settings *JSONRPCServerSettings, handler Handler) (*JSONRPCServer, error) {
	routeGroups := []*http.RouteGroup{
		{
			// these handlers will be executed for all routes in the group
			Handlers: []http.Handler{
				Cors(settings.Cors, false),
			},
			Routes: []*http.Route{
				{
					Pattern: fmt.Sprintf("^%s$", settings.Path),
					Handlers: []http.Handler{
						ExtractJSONRequest,
						JSONRPC(handler),
					},
				},
				{
					Pattern: "^.*$",
					Handlers: []http.Handler{
						NotFound,
					},
				},
			},
		},
	}

	httpServerSettings := &http.HTTPServerSettings{
		TLS:         settings.TLS,
		BindAddress: settings.BindAddress,
	}

	if httpServer, err := http.MakeHTTPServer(httpServerSettings, routeGroups); err != nil {
		return nil, err
	} else {
		return &JSONRPCServer{
			settings: settings,
			server:   httpServer,
		}, nil
	}
}

func (s *JSONRPCServer) HTTPServer() *http.HTTPServer {
	return s.server
}

func (s *JSONRPCServer) Start() error {
	return s.server.Start()
}

func (s *JSONRPCServer) Stop() error {
	return s.server.Stop()
}
