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
	"github.com/iris-gateway/eps"
	"github.com/iris-gateway/eps/helpers"
	"github.com/urfave/cli"
	"os"
	"os/signal"
	"syscall"
)

func openChannels(settings *eps.Settings) []eps.Channel {

	channels, err := helpers.InitializeChannels(settings)

	if err != nil {
		eps.Log.Fatal(err)
	} else {
		for _, channel := range channels {
			if err := channel.Open(); err != nil {
				eps.Log.Fatal(err)
			}
		}
	}
	return channels
}

func closeChannels(channels []eps.Channel) {
	for _, channel := range channels {
		if err := channel.Close(); err != nil {
			eps.Log.Error(err)
		}
	}
}

func Server(settings *eps.Settings) ([]cli.Command, error) {

	return []cli.Command{
		{
			Name:    "server",
			Aliases: []string{"s"},
			Flags:   []cli.Flag{},
			Usage:   "Run the server.",
			Subcommands: []cli.Command{
				{
					Name: "run",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "messages",
							Value: "",
							Usage: "optional: a YAML file with messages to send into the broker",
						},
					}, Usage: "Run the server.",
					Action: func(c *cli.Context) error {
						eps.Log.Info("Opening all channels...")

						channels := openChannels(settings)

						// we wait for CTRL-C / Interrupt
						sigchan := make(chan os.Signal, 1)
						signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

						eps.Log.Info("Waiting for CTRL-C...")

						<-sigchan

						eps.Log.Info("Stopping channels...")

						closeChannels(channels)

						return nil
					},
				},
			},
		},
	}, nil
}
