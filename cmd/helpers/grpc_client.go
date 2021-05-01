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
	"github.com/urfave/cli"
)

func Client(definitions *eps.Definitions, settings *eps.Settings) ([]cli.Command, error) {

	return []cli.Command{
		/*
			{
				Name:    "client",
				Aliases: []string{"s"},
				Flags:   []cli.Flag{},
				Usage:   "gRPC client functionality.",
				Subcommands: []cli.Command{
					{
						Name:  "test",
						Flags: []cli.Flag{},
						Usage: "Test the server connection.",
						Action: func(c *cli.Context) error {
							eps.Log.Info("Starting the client...")

							if settings.GRPCClient == nil {
								eps.Log.Fatalf("gRPC client settings missing!")
							}

							client, err := grpc.MakeClient(settings.GRPCClient, "localhost:4444", "grpc-server")

							if err != nil {
								eps.Log.Fatal(err)
							}

							if err := client.Connect(); err != nil {
								eps.Log.Fatal(err)
							}

							if err := client.SendMessage(); err != nil {
								eps.Log.Fatal(err)
							}

							// we wait for CTRL-C / Interrupt
							sigchan := make(chan os.Signal, 1)
							signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

							eps.Log.Info("Waiting for CTRL-C...")

							<-sigchan

							eps.Log.Info("Stopping client...")

							return nil
						},
					},
				},
			},
		*/
	}, nil
}
