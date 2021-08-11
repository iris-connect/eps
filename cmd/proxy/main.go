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

package main

import (
	"github.com/iris-connect/eps"
	cmdHelpers "github.com/iris-connect/eps/cmd/helpers"
	"github.com/iris-connect/eps/metrics"
	"github.com/iris-connect/eps/proxy"
	"github.com/iris-connect/eps/proxy/helpers"
	"github.com/urfave/cli"
	"os"
	"os/signal"
	"syscall"
)

func CLI(settings *proxy.Settings) {

	var err error

	app := cli.NewApp()
	app.Name = "Proxy Server"
	app.Usage = "Run all proxy server commands"
	app.Flags = cmdHelpers.CommonFlags

	bareCommands := []cli.Command{
		{
			Name:  "run",
			Usage: "Run the public or private proxy",
			Subcommands: []cli.Command{
				{
					Name:  "public",
					Flags: []cli.Flag{},
					Usage: "Run the public TLS proxy.",
					Action: func(c *cli.Context) error {
						eps.Log.Info("Starting public proxy...")

						if settings.Public == nil {
							eps.Log.Fatal("Public settings undefined!")
						}

						server, err := proxy.MakePublicServer(settings.Public)

						if err != nil {
							eps.Log.Fatal(err)
						}

						if err := server.Start(); err != nil {
							eps.Log.Fatal(err)
						}

						metricsServer := metrics.MakePrometheusMetricsServer(settings.Metrics)

						// we wait for CTRL-C / Interrupt
						sigchan := make(chan os.Signal, 1)
						signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

						eps.Log.Info("Waiting for CTRL-C...")

						<-sigchan

						eps.Log.Info("Stopping proxy...")

						if err := server.Stop(); err != nil {
							eps.Log.Fatal(err)
						}

						if metricsServer != nil {
							if err := metricsServer.Stop(); err != nil {
								eps.Log.Fatal(err)
							}
						}

						return nil
					},
				},
				{
					Name:  "private",
					Flags: []cli.Flag{},
					Usage: "Run the private TLS proxy.",
					Action: func(c *cli.Context) error {
						eps.Log.Info("Starting private proxy...")

						if settings.Private == nil {
							eps.Log.Fatal("Private settings undefined!")
						}

						server, err := proxy.MakePrivateServer(settings.Private)

						if err != nil {
							eps.Log.Fatal(err)
						}

						if err := server.Start(); err != nil {
							eps.Log.Fatal(err)
						}

						metricsServer := metrics.MakePrometheusMetricsServer(settings.Metrics)

						// we wait for CTRL-C / Interrupt
						sigchan := make(chan os.Signal, 1)
						signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

						eps.Log.Info("Waiting for CTRL-C...")

						<-sigchan

						eps.Log.Info("Stopping proxy...")

						if err := server.Stop(); err != nil {
							eps.Log.Fatal(err)
						}

						if metricsServer != nil {
							if err := metricsServer.Stop(); err != nil {
								eps.Log.Fatal(err)
							}
						}

						return nil
					},
				},
			},
		},
	}

	app.Commands = cmdHelpers.Decorate(bareCommands, cmdHelpers.InitCLI, "PRO")

	err = app.Run(os.Args)

	if err != nil {
		eps.Log.Error(err)
	}

}

func main() {
	if settings, err := helpers.Settings(helpers.SettingsPaths()); err != nil {
		eps.Log.Error(err)
		return
	} else {
		CLI(settings)
	}
}
