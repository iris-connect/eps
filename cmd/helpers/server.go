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

package helpers

import (
	"github.com/iris-connect/eps"
	"github.com/iris-connect/eps/helpers"
	"github.com/iris-connect/eps/metrics"
	"github.com/urfave/cli"
	"os"
	"os/signal"
	"syscall"
)

func Server(settings *eps.Settings) ([]cli.Command, error) {

	return []cli.Command{
		{
			Name:    "server",
			Aliases: []string{"s"},
			Flags:   []cli.Flag{},
			Usage:   "Server-related commands.",
			Subcommands: []cli.Command{
				{
					Name:  "run",
					Flags: []cli.Flag{},
					Usage: "Run the EPS server.",
					Action: func(c *cli.Context) error {
						eps.Log.Info("Opening all channels...")

						directory, err := helpers.InitializeDirectory(settings)

						if err != nil {
							eps.Log.Fatal(err)
						}

						broker, err := helpers.InitializeMessageBroker(settings, directory)

						if err != nil {
							eps.Log.Fatal(err)
						}

						channels, err := helpers.OpenChannels(broker, directory, settings)

						if err != nil {
							eps.Log.Fatal(err)
						}

						// we wait for CTRL-C / Interrupt
						sigchan := make(chan os.Signal, 1)
						signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

						metricsServer := metrics.MakePrometheusMetricsServer(settings.Metrics)

						eps.Log.Info("Waiting for CTRL-C...")

						<-sigchan

						eps.Log.Info("Stopping channels...")

						// errors occuring within CloseChannels get logged automatically...
						helpers.CloseChannels(channels)

						if metricsServer != nil {
							if err := metricsServer.Stop(); err != nil {
								eps.Log.Error(err)
							}
						}

						return nil
					},
				},
			},
		},
	}, nil
}
