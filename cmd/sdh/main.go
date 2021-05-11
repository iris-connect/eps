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
	"encoding/json"
	"fmt"
	"github.com/iris-gateway/eps"
	"github.com/iris-gateway/eps/sd"
	"github.com/iris-gateway/eps/sd/helpers"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
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

func CLI(settings *sd.SigningSettings) {

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
	app.Name = "Service Directory Helper"
	app.Usage = "Sign and submit service directory entries"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "level",
			Value: "info",
			Usage: "The desired log level",
		},
	}

	bareCommands := []cli.Command{
		{
			Name:  "sign",
			Flags: []cli.Flag{},
			Usage: "Sign a JSON entry",
			Action: func(c *cli.Context) error {

				filename := c.Args().Get(0)

				if filename == "" {
					eps.Log.Fatal("please specify a filename")
				}

				jsonBytes, err := ioutil.ReadFile(filename)

				if err != nil {
					eps.Log.Fatal(err)
				}

				var jsonData map[string]interface{}

				if err := json.Unmarshal(jsonBytes, &jsonData); err != nil {
					eps.Log.Fatal(err)
				}

				certificate, err := helpers.LoadCertificate(settings.CertificateFile, true)

				if err != nil {
					eps.Log.Fatal(err)
				}

				rootCertificate, err := helpers.LoadCertificate(settings.CACertificateFile, false)

				if err != nil {
					eps.Log.Fatal(err)
				}

				// we ensure the certificate is valid for signing
				if err := helpers.VerifyCertificate(certificate, rootCertificate, settings.Name); err != nil {
					eps.Log.Fatal(err)
				}

				key, err := helpers.LoadPrivateKey(settings.KeyFile)

				if err != nil {
					eps.Log.Fatal(err)
				}

				signedData, err := helpers.Sign(jsonData, key, certificate)

				if err != nil {
					eps.Log.Fatal(err)
				}

				signedDataBytes, err := json.Marshal(signedData)

				if err != nil {
					eps.Log.Fatal(err)
				}

				fmt.Println(string(signedDataBytes))

				loadedSignedData, err := helpers.LoadSignedData(signedDataBytes)

				if err != nil {
					eps.Log.Fatal(err)
				}

				if ok, err := helpers.Verify(loadedSignedData, rootCertificate, settings.Name); err != nil {
					eps.Log.Fatal(err)
				} else if !ok {
					eps.Log.Fatal("Signature is not valid!")
				}

				return nil
			},
		},
		{
			Name:  "verify",
			Flags: []cli.Flag{},
			Usage: "Verify a JSON entry",
			Action: func(c *cli.Context) error {

				filename := c.Args().Get(0)

				if filename == "" {
					eps.Log.Fatal("please specify a filename")
				}

				name := c.Args().Get(1)

				if name == "" {
					eps.Log.Fatal("please specify a name")
				}

				jsonBytes, err := ioutil.ReadFile(filename)

				if err != nil {
					eps.Log.Fatal(err)
				}

				var signedData *helpers.SignedData

				if err := json.Unmarshal(jsonBytes, &signedData); err != nil {
					eps.Log.Fatal(err)
				}

				rootCertificate, err := helpers.LoadCertificate(settings.CACertificateFile, false)

				if err != nil {
					eps.Log.Fatal(err)
				}

				if ok, err := helpers.Verify(signedData, rootCertificate, name); err != nil {
					eps.Log.Fatal(err)
				} else if !ok {
					eps.Log.Fatal("Signature is not valid!")
				}

				return nil
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
	if settings, err := helpers.SigningSettings(helpers.SettingsPaths()); err != nil {
		eps.Log.Error(err)
		return
	} else {
		CLI(settings)
	}
}
