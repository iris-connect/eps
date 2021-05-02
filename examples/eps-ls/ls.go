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
	"github.com/iris-gateway/eps/jsonrpc"
	"github.com/iris-gateway/eps"
	"os"
	"os/signal"
	"syscall"
)

func handler(context *jsonrpc.Context) *jsonrpc.Response {
	return context.Result("that's ok")
}

func main() {

	settings := &jsonrpc.JSONRPCServerSettings{
		BindAddress: "localhost:6666",
	}

	if server, err := jsonrpc.MakeJSONRPCServer(settings, handler); err != nil {
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
