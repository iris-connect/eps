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
	"github.com/iris-gateway/eps/http"
	"github.com/iris-gateway/eps/tls"
	"github.com/iris-gateway/eps"
	"os"
	"os/signal"
	"syscall"
)

func handler(context *http.Context) {
	context.JSON(200, map[string]interface{}{"message" : "success"})
}

func main() {

	bindAddress := os.Getenv("IS_BIND_ADDRESS")

	if bindAddress == "" {
		// this is just a test server so we bind it to 0.0.0.0 by default to
		// make testing and development easier. We'll never bind production
		// servers to 0.0.0.0 by default for security reasons.
		bindAddress = "0.0.0.0:8888"
	}

	settings := &http.HTTPServerSettings{
		BindAddress: bindAddress,
		TLS: &tls.TLSSettings{
			CACertificateFile: "settings/dev/certs/root.crt",
			CertificateFile: "settings/dev/certs/internal-server.crt",
			KeyFile: "settings/dev/certs/internal-server.key",
			VerifyClient: false,
		},
	}

	routeGroups := []*http.RouteGroup{
		{
			// these handlers will be executed for all routes in the group
			Handlers: []http.Handler{},
			Routes: []*http.Route{
				{
					Pattern: "^.*$",
					Handlers: []http.Handler{
						handler,
					},
				},
			},
		},
	}

	if server, err := http.MakeHTTPServer(settings, routeGroups); err != nil {
		eps.Log.Fatal(err)
	} else {
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
