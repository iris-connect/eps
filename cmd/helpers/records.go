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
	epsForms "github.com/iris-gateway/eps/forms"
	"github.com/iris-gateway/eps/helpers"
	"github.com/kiprotect/go-helpers/forms"
	"github.com/urfave/cli"
	"io/ioutil"
	"time"
)

var RecordsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "records",
			Validators: []forms.Validator{
				forms.IsList{
					Validators: []forms.Validator{
						forms.IsStringMap{
							Form: &epsForms.ChangeRecordForm,
						},
					},
				},
			},
		},
	},
}

type Records struct {
	Records []*eps.ChangeRecord `json:"records"`
}

func submitRecords(c *cli.Context, settings *eps.Settings) error {

	reset := c.Bool("reset")

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

	records := &Records{}
	var rawRecords map[string]interface{}

	if err := json.Unmarshal(jsonBytes, &rawRecords); err != nil {
		eps.Log.Fatal(err)
	}

	if params, err := RecordsForm.Validate(rawRecords); err != nil {
		eps.Log.Fatal(err)
	} else if RecordsForm.Coerce(records, params); err != nil {
		eps.Log.Fatal(err)
	}

	if err := submitChangeRecords(records.Records, settings, reset); err != nil {
		eps.Log.Fatal(err)
	}

	return nil

}

func submitChangeRecords(changeRecords []*eps.ChangeRecord, settings *eps.Settings, reset bool) error {

	directory, err := helpers.InitializeDirectory(settings)

	if err != nil {
		eps.Log.Fatal(err)
	}

	writableDirectory, ok := directory.(eps.WritableDirectory)

	if !ok {
		eps.Log.Fatalf("not a writable service directory")
	}

	certificate, err := helpers.LoadCertificate(settings.Signing.CertificateFile, true)

	if err != nil {
		eps.Log.Fatal(err)
	}

	rootCertificate, err := helpers.LoadCertificate(settings.Signing.CACertificateFile, false)

	if err != nil {
		eps.Log.Fatal(err)
	}

	// we ensure the certificate is valid for signing
	if err := helpers.VerifyCertificate(certificate, rootCertificate, settings.Name); err != nil {
		eps.Log.Fatal(err)
	}

	key, err := helpers.LoadPrivateKey(settings.Signing.KeyFile)

	if err != nil {
		eps.Log.Fatal(err)
	}

	lastRecord, err := writableDirectory.Tip()

	if err != nil {
		eps.Log.Fatal(err)
	}

	var parentHash string

	if lastRecord != nil && !reset {
		parentHash = lastRecord.Hash
	}

	signedChangeRecords := make([]*eps.SignedChangeRecord, 0)

	for _, changeRecord := range changeRecords {

		changeRecord.CreatedAt = eps.HashableTime{time.Now()}

		signedChangeRecord := &eps.SignedChangeRecord{
			ParentHash: parentHash,
			Record:     changeRecord,
		}

		if err := helpers.CalculateRecordHash(signedChangeRecord); err != nil {
			eps.Log.Fatal(err)
		}

		eps.Log.Info(signedChangeRecord.Hash)

		signedData, err := helpers.Sign(signedChangeRecord, key, certificate)

		if err != nil {
			eps.Log.Fatal(err)
		}

		signedChangeRecord.Signature = signedData.Signature
		signedChangeRecords = append(signedChangeRecords, signedChangeRecord)
		parentHash = signedChangeRecord.Hash
	}

	if err := writableDirectory.Submit(signedChangeRecords); err != nil {
		eps.Log.Fatal(err)
	}

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

	certificate, err := helpers.LoadCertificate(settings.Signing.CertificateFile, true)

	if err != nil {
		eps.Log.Fatal(err)
	}

	rootCertificate, err := helpers.LoadCertificate(settings.Signing.CACertificateFile, false)

	if err != nil {
		eps.Log.Fatal(err)
	}

	// we ensure the certificate is valid for signing
	if err := helpers.VerifyCertificate(certificate, rootCertificate, settings.Name); err != nil {
		eps.Log.Fatal(err)
	}

	key, err := helpers.LoadPrivateKey(settings.Signing.KeyFile)

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

	rootCertificate, err := helpers.LoadCertificate(settings.Signing.CACertificateFile, false)

	if err != nil {
		eps.Log.Fatal(err)
	}

	if ok, err := helpers.Verify(signedData, rootCertificate, name); err != nil {
		eps.Log.Fatal(err)
	} else if !ok {
		eps.Log.Fatal("Signature is not valid!")
	} else {
		eps.Log.Info("Signature is ok!")
	}
	return nil
}
func RecordsCommands(settings *eps.Settings) ([]cli.Command, error) {

	return []cli.Command{
		{
			Name:    "sd",
			Aliases: []string{"s"},
			Flags:   []cli.Flag{},
			Usage:   "Manage service-directory records.",
			Subcommands: []cli.Command{
				{
					Name: "submit-records",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "reset",
							Usage: "reset the remote records (dangerous)",
						},
					},
					Usage:  "Submit several records at once",
					Action: func(c *cli.Context) error { return submitRecords(c, settings) },
				},
				{
					Name:   "sign",
					Flags:  []cli.Flag{},
					Usage:  "Sign a change record",
					Action: func(c *cli.Context) error { return sign(c, settings) },
				},
				{
					Name:   "verify",
					Flags:  []cli.Flag{},
					Usage:  "Verify a change record",
					Action: func(c *cli.Context) error { return verify(c, settings) },
				},
			},
		},
	}, nil
}
