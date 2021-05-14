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

var DirectoryForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "entries",
			Validators: []forms.Validator{
				forms.IsList{
					Validators: []forms.Validator{
						forms.IsStringMap{
							Form: &epsForms.DirectoryEntryForm,
						},
					},
				},
			},
		},
	},
}

type Directory struct {
	Entries []*eps.DirectoryEntry `json:"entries"`
}

func submitDirectory(c *cli.Context, settings *eps.Settings) error {

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

	var directory *Directory

	if err := json.Unmarshal(jsonBytes, &directory); err != nil {
		eps.Log.Fatal(err)
	}

	submitRecord := func(changeRecord *eps.ChangeRecord) {
		if err := submitChangeRecord(changeRecord, settings); err != nil {
			eps.Log.Fatal(err)
		}
	}

	for i, entry := range directory.Entries {
		changeRecord := &eps.ChangeRecord{}
		changeRecord.Name = entry.Name
		if len(entry.Settings) > 0 {
			eps.Log.Infof("Submitting settings for entry %d...", i)
			changeRecord.Section = "settings"
			changeRecord.Data = entry.Settings
			submitRecord(changeRecord)
		}
		if len(entry.Channels) > 0 {
			eps.Log.Infof("Submitting channels for entry %d...", i)
			changeRecord.Section = "channels"
			changeRecord.Data = entry.Channels
			submitRecord(changeRecord)
		}
		if len(entry.Services) > 0 {
			eps.Log.Infof("Submitting services for entry %d...", i)
			changeRecord.Section = "services"
			changeRecord.Data = entry.Services
			submitRecord(changeRecord)
		}
		if len(entry.Certificates) > 0 {
			eps.Log.Infof("Submitting certificates for entry %d...", i)
			changeRecord.Section = "certificates"
			changeRecord.Data = entry.Certificates
			submitRecord(changeRecord)
		}
		if len(entry.Settings) > 0 {
			eps.Log.Infof("Submitting settings for entry %d...", i)
			changeRecord.Section = "settings"
			changeRecord.Data = entry.Settings
			submitRecord(changeRecord)
		}
	}

	return nil

}

func submitChangeRecord(changeRecord *eps.ChangeRecord, settings *eps.Settings) error {

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

	var lastPosition int64

	if lastRecord != nil {
		lastPosition = lastRecord.Position
	}

	changeRecord.CreatedAt = eps.HashableTime{time.Now()}

	signedChangeRecord := &eps.SignedChangeRecord{
		Position: lastPosition + 1,
		Record:   changeRecord,
	}

	lastHash := ""

	if lastRecord != nil {
		lastHash = lastRecord.Hash
	}

	if newHash, err := helpers.CalculateHash(signedChangeRecord, lastHash); err != nil {
		eps.Log.Fatal(err)
	} else {
		signedChangeRecord.Hash = newHash
	}

	signedData, err := helpers.Sign(signedChangeRecord, key, certificate)
	signedChangeRecord.Signature = signedData.Signature

	if err := writableDirectory.Submit(signedChangeRecord); err != nil {
		eps.Log.Fatal(err)
	}

	eps.Log.Info("Successfully submitted record!")

	return nil

}

func submitRecord(c *cli.Context, settings *eps.Settings) error {
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

	var changeRecord *eps.ChangeRecord

	if err := json.Unmarshal(jsonBytes, &changeRecord); err != nil {
		eps.Log.Fatal(err)
	}

	return submitChangeRecord(changeRecord, settings)
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
func Records(settings *eps.Settings) ([]cli.Command, error) {

	return []cli.Command{
		{
			Name:    "sd",
			Aliases: []string{"s"},
			Flags:   []cli.Flag{},
			Usage:   "Manage service-directory records.",
			Subcommands: []cli.Command{
				{
					Name:   "submit-record",
					Flags:  []cli.Flag{},
					Usage:  "Submit a JSON change record",
					Action: func(c *cli.Context) error { return submitRecord(c, settings) },
				},
				{
					Name:   "submit-directory",
					Flags:  []cli.Flag{},
					Usage:  "Submit a full service directory",
					Action: func(c *cli.Context) error { return submitDirectory(c, settings) },
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
