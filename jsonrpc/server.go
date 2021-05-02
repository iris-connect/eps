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
	"github.com/iris-gateway/eps"
	"github.com/iris-gateway/eps/http"
)

type JSONRPCServer struct {
	settings *JSONRPCServerSettings
	server   *http.HTTPServer
}

func JSONRPC(c *http.Context) {
	requestData := c.Get("requestData").(*RequestData)
	c.JSON(200, requestData)
}

func MakeJSONRPCServer(settings *JSONRPCServerSettings) (*JSONRPCServer, error) {
	routeGroups := []*http.RouteGroup{
		{
			// these handlers will be executed for all routes in the group
			Handlers: []http.Handler{
				Cors(settings.Cors, false),
			},
			Routes: []*http.Route{
				{
					Pattern: "^/jsonrpc$",
					Handlers: []http.Handler{
						ExtractJSONRequestData,
						JSONRPC,
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

func (s *JSONRPCServer) Start() error {
	return s.server.Start()
}

func (s *JSONRPCServer) Stop() error {
	eps.Log.Debugf("Stopping down JSONRPC server...")
	return s.server.Stop()
}
