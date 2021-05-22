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
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"github.com/iris-connect/eps"
	epsForms "github.com/iris-connect/eps/forms"
	"github.com/kiprotect/go-helpers/forms"
)

func InitializeDirectory(settings *eps.Settings) (eps.Directory, error) {
	definition := settings.Definitions.DirectoryDefinitions[settings.Directory.Type]
	return definition.Maker(settings.Name, settings.Directory.Settings)
}

// Integrates a record into the directory
func IntegrateChangeRecord(record *eps.SignedChangeRecord, entry *eps.DirectoryEntry) error {

	config := map[string]interface{}{
		record.Record.Section: record.Record.Data,
	}

	// we directly coerce the updated settings into the entry
	if err := epsForms.DirectoryEntryForm.Coerce(entry, config); err != nil {
		return err
	} else {
		if entry.Records == nil {
			entry.Records = make([]*eps.SignedChangeRecord, 0)
		}
		// we append the change record to the entry for audit logging purposes
		entry.Records = append(entry.Records, record)
	}
	return nil
}

type CertificatesList struct {
	Certificates []*eps.OperatorCertificate `json:"certificates"`
}

var CertificatesListForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "certificates",
			Validators: []forms.Validator{
				forms.IsOptional{Default: []interface{}{}},
				forms.IsList{
					Validators: []forms.Validator{
						forms.IsStringMap{
							Form: &epsForms.OperatorCertificateForm,
						},
					},
				},
			},
		},
	},
}

func GetRecordFingerprint(records []*eps.SignedChangeRecord, name, keyUsage string) string {
	lastFingerprint := ""
	for _, record := range records {
		if record.Record.Section != "certificates" || record.Record.Name != name {
			continue
		}
		if params, err := CertificatesListForm.Validate(map[string]interface{}{"certificates": record.Record.Data}); err != nil {
			eps.Log.Error(err)
			continue
		} else {
			certificatesList := &CertificatesList{}
			if err := CertificatesListForm.Coerce(certificatesList, params); err != nil {
				eps.Log.Error(err)
				continue
			}
			for _, certificate := range certificatesList.Certificates {
				if certificate.KeyUsage == keyUsage {
					lastFingerprint = certificate.Fingerprint
				}
			}
		}
	}
	return lastFingerprint
}

func VerifyRecordHash(record *eps.SignedChangeRecord) (bool, error) {

	submittedHash := record.Hash

	err := CalculateRecordHash(record)

	if err != nil {
		return false, err
	}

	if submittedHash != record.Hash {
		return false, nil
	}

	return true, nil
}

func VerifyRecord(record *eps.SignedChangeRecord, verifiedRecords []*eps.SignedChangeRecord, rootCerts []*x509.Certificate) (bool, error) {
	signedData := &eps.SignedData{
		Data:      record,
		Signature: record.Signature,
	}

	if ok, err := VerifyRecordHash(record); err != nil {
		return false, err
	} else if !ok {
		return false, fmt.Errorf("invalid hash value")
	}

	// we temporarily remove the signature from the signed record
	signature := record.Signature
	record.Signature = nil
	defer func() { record.Signature = signature }()

	cert, err := LoadCertificateFromString(signature.Certificate, true)

	if err != nil {
		return false, err
	}

	subjectInfo, err := GetSubjectInfo(cert)

	if err != nil {
		return false, err
	}

	fingerprint := GetRecordFingerprint(verifiedRecords, subjectInfo.Name, "signing")

	if fingerprint == "" {
		admin := false
		for _, group := range subjectInfo.Groups {
			if group == "sd-admin" {
				// service directory admins can upload its own certificate info
				// (but only if that info doesn't exist yet)
				admin = true
				break
			}
		}
		// only a service directory admin can proceed without fingerprint validation
		if !admin {
			return false, nil
		}
	} else if !VerifyFingerprint(cert, fingerprint) {
		// the fingerprint does not match the one we have on record
		return false, nil
	}

	return Verify(signedData, rootCerts, "")
}

func CalculateRecordHash(record *eps.SignedChangeRecord) error {

	// we always reset the hash before calculating the new one
	record.Hash = ""

	hash, err := StructuredHash(record.Record)

	if err != nil {
		return err
	}

	record.Hash = hex.EncodeToString(hash[:])

	return nil

}
