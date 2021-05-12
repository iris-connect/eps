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
	"encoding/json"
	"fmt"
	"github.com/iris-gateway/eps"
	"github.com/iris-gateway/eps/helpers"
	"github.com/urfave/cli"
	"io/ioutil"
	"time"
)

func submit(c *cli.Context, settings *eps.Settings) error {
	if settings.Signing == nil {
		eps.Log.Fatalf("Signing settings undefined!")
	}

	directory, err := helpers.InitializeDirectory(settings)

	if err != nil {
		eps.Log.Fatal(err)
	}

	writableDirectory, ok := directory.(eps.WritableDirectory)

	if !ok {
		eps.Log.Fatalf("not a writable service directory")
	}

	filename := c.Args().Get(0)

	if filename == "" {
		eps.Log.Fatal("please specify a filename")
	}

	jsonBytes, err := ioutil.ReadFile(filename)

	if err != nil {
		eps.Log.Fatal(err)
	}

	var changeRecord *eps.ChangeRecord

	if err := json.Unmarshal(jsonBytes, &changeRecord); err != nil {
		eps.Log.Fatal(err)
	}

	certificate, err := eps.LoadCertificate(settings.Signing.CertificateFile, true)

	if err != nil {
		eps.Log.Fatal(err)
	}

	rootCertificate, err := eps.LoadCertificate(settings.Signing.CACertificateFile, false)

	if err != nil {
		eps.Log.Fatal(err)
	}

	// we ensure the certificate is valid for signing
	if err := eps.VerifyCertificate(certificate, rootCertificate, settings.Name); err != nil {
		eps.Log.Fatal(err)
	}

	key, err := eps.LoadPrivateKey(settings.Signing.KeyFile)

	if err != nil {
		eps.Log.Fatal(err)
	}

	lastRecord, err := writableDirectory.Tip()

	if err != nil {
		eps.Log.Fatal(err)
	}

	var lastPosition int64

	if lastRecord != nil {
		lastPosition = lastRecord.Position
	}

	changeRecord.CreatedAt = time.Now()

	signedChangeRecord := &eps.SignedChangeRecord{
		Position: lastPosition + 1,
		Record:   changeRecord,
	}

	lastHash := ""

	if lastRecord != nil {
		lastHash = lastRecord.Hash
	}

	if newHash, err := signedChangeRecord.CalculateHash(lastHash); err != nil {
		eps.Log.Fatal(err)
	} else {
		signedChangeRecord.Hash = newHash
	}

	signedData, err := eps.Sign(signedChangeRecord, key, certificate)
	signedChangeRecord.Signature = signedData.Signature

	data, _ := json.Marshal(signedChangeRecord)

	fmt.Println(string(data))

	if err := writableDirectory.Submit(signedChangeRecord); err != nil {
		eps.Log.Fatal(err)
	}

	eps.Log.Info("Successfully submitted record!")

	return nil
}

func sign(c *cli.Context, settings *eps.Settings) error {
	if settings.Signing == nil {
		eps.Log.Fatalf("Signing settings undefined!")
	}

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

	certificate, err := eps.LoadCertificate(settings.Signing.CertificateFile, true)

	if err != nil {
		eps.Log.Fatal(err)
	}

	rootCertificate, err := eps.LoadCertificate(settings.Signing.CACertificateFile, false)

	if err != nil {
		eps.Log.Fatal(err)
	}

	// we ensure the certificate is valid for signing
	if err := eps.VerifyCertificate(certificate, rootCertificate, settings.Name); err != nil {
		eps.Log.Fatal(err)
	}

	key, err := eps.LoadPrivateKey(settings.Signing.KeyFile)

	if err != nil {
		eps.Log.Fatal(err)
	}

	signedData, err := eps.Sign(jsonData, key, certificate)

	if err != nil {
		eps.Log.Fatal(err)
	}

	signedDataBytes, err := json.Marshal(signedData)

	if err != nil {
		eps.Log.Fatal(err)
	}

	fmt.Println(string(signedDataBytes))

	loadedSignedData, err := eps.LoadSignedData(signedDataBytes)

	if err != nil {
		eps.Log.Fatal(err)
	}

	if ok, err := eps.Verify(loadedSignedData, rootCertificate, settings.Name); err != nil {
		eps.Log.Fatal(err)
	} else if !ok {
		eps.Log.Fatal("Signature is not valid!")
	}

	return nil
}

func verify(c *cli.Context, settings *eps.Settings) error {

	if settings.Signing == nil {
		eps.Log.Fatalf("Signing settings undefined!")
	}

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

	var signedData *eps.SignedData

	if err := json.Unmarshal(jsonBytes, &signedData); err != nil {
		eps.Log.Fatal(err)
	}

	rootCertificate, err := eps.LoadCertificate(settings.Signing.CACertificateFile, false)

	if err != nil {
		eps.Log.Fatal(err)
	}

	if ok, err := eps.Verify(signedData, rootCertificate, name); err != nil {
		eps.Log.Fatal(err)
	} else if !ok {
		eps.Log.Fatal("Signature is not valid!")
	} else {
		eps.Log.Info("Signature is ok!")
	}
	return nil
}
func Records(settings *eps.Settings) ([]cli.Command, error) {

	return []cli.Command{
		{
			Name:    "records",
			Aliases: []string{"s"},
			Flags:   []cli.Flag{},
			Usage:   "Manage service-directory records.",
			Subcommands: []cli.Command{
				{
					Name:   "submit",
					Flags:  []cli.Flag{},
					Usage:  "Submit a JSON entry",
					Action: func(c *cli.Context) error { return submit(c, settings) },
				},
				{
					Name:   "sign",
					Flags:  []cli.Flag{},
					Usage:  "Sign a JSON entry",
					Action: func(c *cli.Context) error { return sign(c, settings) },
				},
				{
					Name:   "verify",
					Flags:  []cli.Flag{},
					Usage:  "Verify a JSON entry",
					Action: func(c *cli.Context) error { return verify(c, settings) },
				},
			},
		},
	}, nil
}
