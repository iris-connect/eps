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
	"encoding/json"
	"github.com/iris-connect/eps"
	"github.com/iris-connect/eps/http"
	"github.com/iris-connect/eps/tls"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
)

func handler(context *http.Context) {

	ise := func(){
		context.JSON(500, map[string]interface{}{"message": "internal server error"})
	}

	eps.Log.Debugf("Received request with path '%s' and query' %s'", context.Request.URL.Path, context.Request.URL.RawQuery)

	if context.Request.Method == "POST" {

		body, err := ioutil.ReadAll(context.Request.Body)

		if err != nil {
			ise()
			return
		}

		var jsonData map[string]interface{}
		if err := json.Unmarshal(body, &jsonData); err != nil {
			ise()
			return
		}
		context.JSON(200, jsonData)
	}
	context.JSON(200, map[string]interface{}{"message" : "success"})
}

func main() {

	bindAddress := os.Getenv("IS_BIND_ADDRESS")
	useTls := os.Getenv("USE_TLS")

	if bindAddress == "" {
		// this is just a test server so we bind it to 0.0.0.0 by default to
		// make testing and development easier. We'll never bind production
		// servers to 0.0.0.0 by default for security reasons.
		bindAddress = "0.0.0.0:8888"
	}

	var tlsSettings *tls.TLSSettings

	if useTls != "" {
		tlsSettings = &tls.TLSSettings{
			CACertificateFiles: []string{"settings/dev/certs/root.crt"},
			CertificateFile: "settings/dev/certs/internal-server.crt",
			KeyFile: "settings/dev/certs/internal-server.key",
			VerifyClient: false,
		}

	}

	settings := &http.HTTPServerSettings{
		BindAddress: bindAddress,
		TLS: tlsSettings,
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
