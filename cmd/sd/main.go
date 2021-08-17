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
	"github.com/iris-connect/eps/definitions"
	"github.com/iris-connect/eps/metrics"
	"github.com/iris-connect/eps/sd"
	"github.com/iris-connect/eps/sd/helpers"
	"github.com/urfave/cli"
	"os"
	"os/signal"
	"syscall"
)

func CLI(settings *sd.Settings) {

	var err error

	app := cli.NewApp()
	app.Name = "Service Directory"
	app.Usage = "Run all service directory commands"
	app.Flags = cmdHelpers.CommonFlags

	bareCommands := append([]cli.Command{
		{
			Name:  "run",
			Flags: []cli.Flag{},
			Usage: "Run the service directory.",
			Action: func(c *cli.Context) error {
				eps.Log.Info("Starting the service directory...")
				server, err := sd.MakeServer(settings)

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

				eps.Log.Info("Stopping directory...")

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
	}, cmdHelpers.CommonCommands...)

	app.Commands = cmdHelpers.Decorate(bareCommands, cmdHelpers.InitCLI, "SD_")

	err = app.Run(os.Args)

	if err != nil {
		eps.Log.Error(err)
	}

}

func main() {
	if settings, err := helpers.Settings(helpers.SettingsPaths(), &definitions.Default); err != nil {
		eps.Log.Error(err)
		return
	} else {
		CLI(settings)
	}
}
