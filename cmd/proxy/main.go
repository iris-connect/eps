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
	"github.com/iris-connect/eps/metrics"
	"github.com/iris-connect/eps/proxy"
	"github.com/iris-connect/eps/proxy/helpers"
	"github.com/urfave/cli"
	"os"
	"os/signal"
	"syscall"
)

type decorator func(f func(c *cli.Context) error) func(c *cli.Context) error

func decorate(commands []cli.Command, decorator decorator) []cli.Command {
	newCommands := make([]cli.Command, len(commands))
	for i, command := range commands {
		if command.Action != nil {
			command.Action = decorator(command.Action.(func(c *cli.Context) error))
		}
		if command.Subcommands != nil {
			command.Subcommands = decorate(command.Subcommands, decorator)
		}
		newCommands[i] = command
	}
	return newCommands
}

func CLI(settings *proxy.Settings) {

	var err error

	init := func(f func(c *cli.Context) error) func(c *cli.Context) error {
		return func(c *cli.Context) error {

			level := c.GlobalString("level")
			logLevel, err := eps.ParseLevel(level)
			if err != nil {
				return err
			}
			eps.Log.SetLevel(logLevel)

			return f(c)
		}
	}

	app := cli.NewApp()
	app.Name = "Proxy Server"
	app.Usage = "Run all proxy server commands"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "level",
			Value: "info",
			Usage: "The desired log level",
		},
	}

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

						// we wait for CTRL-C / Interrupt
						sigchan := make(chan os.Signal, 1)
						signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

						eps.Log.Info("Waiting for CTRL-C...")

						<-sigchan

						eps.Log.Info("Stopping proxy...")

						if err := server.Stop(); err != nil {
							eps.Log.Fatal(err)
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

						// we wait for CTRL-C / Interrupt
						sigchan := make(chan os.Signal, 1)
						signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

						eps.Log.Info("Waiting for CTRL-C...")

						<-sigchan

						eps.Log.Info("Stopping proxy...")

						if err := server.Stop(); err != nil {
							eps.Log.Fatal(err)
						}

						return nil
					},
				},
			},
		},
	}

	app.Commands = decorate(bareCommands, init)

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
		metrics.OpenPrometheusEndpoint()
		CLI(settings)
	}
}
