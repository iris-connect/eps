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

package sd

import (
	"crypto/x509"
	"encoding/asn1"
	"encoding/json"
	"fmt"
	"github.com/iris-gateway/eps"
	epsForms "github.com/iris-gateway/eps/forms"
	"github.com/iris-gateway/eps/helpers"
	"github.com/kiprotect/go-helpers/forms"
	"net/url"
	"sync"
)

const (
	SignedChangeRecordEntry uint8 = 1
)

type RecordDirectorySettings struct {
	DatabaseFile      string `json:"database_file"`
	CACertificateFile string `json:"ca_certificate_file"`
}

type RecordDirectory struct {
	rootCert       *x509.Certificate
	dataStore      helpers.DataStore
	settings       *RecordDirectorySettings
	entries        map[string]*eps.DirectoryEntry
	recordsByHash  map[string]*eps.SignedChangeRecord
	recordChildren map[string][]*eps.SignedChangeRecord
	orderedRecords []*eps.SignedChangeRecord
	mutex          sync.Mutex
}

func MakeRecordDirectory(settings *RecordDirectorySettings) (*RecordDirectory, error) {

	cert, err := helpers.LoadCertificate(settings.CACertificateFile, false)

	if err != nil {
		return nil, err
	}

	f := &RecordDirectory{
		rootCert:       cert,
		orderedRecords: make([]*eps.SignedChangeRecord, 0),
		recordsByHash:  make(map[string]*eps.SignedChangeRecord),
		recordChildren: make(map[string][]*eps.SignedChangeRecord),
		settings:       settings,
		dataStore:      helpers.MakeFileDataStore(settings.DatabaseFile),
	}

	if err := f.dataStore.Init(); err != nil {
		return nil, err
	}

	_, err = f.update()

	return f, err
}

func (f *RecordDirectory) Entry(name string) (*eps.DirectoryEntry, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if entry, ok := f.entries[name]; !ok {
		return nil, nil
	} else {
		return entry, nil
	}
}

func (f *RecordDirectory) AllEntries() ([]*eps.DirectoryEntry, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	entries := make([]*eps.DirectoryEntry, len(f.entries))
	i := 0
	for _, entry := range f.entries {
		entries[i] = entry
		i++
	}
	return entries, nil
}

func (f *RecordDirectory) Entries(*eps.DirectoryQuery) ([]*eps.DirectoryEntry, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return nil, nil
}

type SubjectInfo struct {
	Name     string
	DNSNames []string
	Groups   []string
}

func getSubjectInfo(cert *x509.Certificate) (*SubjectInfo, error) {

	// subject alternative name extension
	id := asn1.ObjectIdentifier{2, 5, 29, 17}

	subjectInfo := &SubjectInfo{
		DNSNames: make([]string, 0),
		Groups:   make([]string, 0),
	}

	for _, extension := range cert.Extensions {
		if !extension.Id.Equal(id) {
			continue
		}
		// we unmarshal the ASN.1 object
		altNames := []asn1.RawValue{}
		if _, err := asn1.Unmarshal(extension.Value, &altNames); err != nil {
			return nil, err
		}
		for _, altName := range altNames {
			if altName.Tag == 6 {
				// tag 6 is URI (don't ask why...)
				if groupUrl, err := url.Parse(string(altName.Bytes)); err != nil {
					return nil, err
				} else {
					switch groupUrl.Scheme {
					case "iris-group":
						if groupUrl.Host != "" {
							subjectInfo.Groups = append(subjectInfo.Groups, groupUrl.Host)
						}
					case "iris-name":
						subjectInfo.Name = groupUrl.Host
					}
				}
			} else if altName.Tag == 2 {
				// tag 2 is DNS (don't ask why...)
				subjectInfo.DNSNames = append(subjectInfo.DNSNames, string(altName.Bytes))
			}
		}
	}
	return subjectInfo, nil

}

// determines whether a subject can append to the service directory
func (f *RecordDirectory) canAppend(record *eps.SignedChangeRecord) (bool, error) {

	cert, err := helpers.LoadCertificateFromString(record.Signature.Certificate, true)

	if err != nil {
		return false, err
	}

	subjectInfo, err := getSubjectInfo(cert)

	if err != nil {
		return false, err
	}

	// operators can edit their own channels and set their own preferences
	if subjectInfo.Name == record.Record.Name && (record.Record.Section == "channels" || record.Record.Section == "preferences") {
		return true, nil
	}

	for _, group := range subjectInfo.Groups {
		if group == "sd-admin" {
			// service directory admins can do everything
			return true, nil
		}
	}

	// we verify the signature of the record
	if ok, err := f.verifySignature(record, f.orderedRecords); err != nil {
		return false, err
	} else if !ok {
		return false, nil
	}

	// everything else is forbidden
	return false, nil
}

// Appends a new record
func (f *RecordDirectory) Append(record *eps.SignedChangeRecord) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if _, err := f.update(); err != nil {
		return err
	}

	if ok, err := f.canAppend(record); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("you cannot append")
	}

	tip, err := f.tip()

	if err != nil {
		return err
	}

	if (tip != nil && record.ParentHash != tip.Hash) || (tip == nil && record.ParentHash != "") {
		return fmt.Errorf("stale append, please try again")
	}

	if ok, err := f.verifyHash(record); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("invalid hash provided, please try again")
	}

	id, err := helpers.RandomID(16)

	if err != nil {
		return err
	}

	rawData, err := json.Marshal(record)

	if err != nil {
		return err
	}

	dataEntry := &helpers.DataEntry{
		Type: SignedChangeRecordEntry,
		ID:   id,
		Data: rawData,
	}

	if err := f.dataStore.Write(dataEntry); err != nil {
		return err
	}

	// we update the store
	if newRecords, err := f.update(); err != nil {
		return err
	} else {
		for _, newRecord := range newRecords {
			if newRecord.Hash == record.Hash {
				return nil
			}
		}
		return fmt.Errorf("new record not found")
	}
}

func (f *RecordDirectory) Update() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	_, err := f.update()
	return err
}

// Returns the latest record
func (f *RecordDirectory) Tip() (*eps.SignedChangeRecord, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.tip()
}

func (f *RecordDirectory) tip() (*eps.SignedChangeRecord, error) {
	if len(f.orderedRecords) == 0 {
		return nil, nil
	}
	return f.orderedRecords[len(f.orderedRecords)-1], nil
}

// Returns all records since a given date
func (f *RecordDirectory) Records(since string) ([]*eps.SignedChangeRecord, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	relevantRecords := make([]*eps.SignedChangeRecord, 0)
	found := false
	if since == "" {
		found = true
	}
	for _, record := range f.orderedRecords {
		if record.Hash == since {
			found = true
		}
		if found {
			relevantRecords = append(relevantRecords, record)
		}
	}
	return relevantRecords, nil
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

func getFingerprint(records []*eps.SignedChangeRecord, name, keyUsage string) string {
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

func (f *RecordDirectory) verifySignature(record *eps.SignedChangeRecord, verifiedRecords []*eps.SignedChangeRecord) (bool, error) {
	signedData := &eps.SignedData{
		Data:      record,
		Signature: record.Signature,
	}

	// we temporarily remove the signature from the signed record
	signature := record.Signature
	record.Signature = nil
	defer func() { record.Signature = signature }()

	cert, err := helpers.LoadCertificateFromString(signature.Certificate, true)

	if err != nil {
		return false, err
	}

	subjectInfo, err := getSubjectInfo(cert)

	if err != nil {
		return false, err
	}

	fingerprint := getFingerprint(verifiedRecords, subjectInfo.Name, "signing")

	if fingerprint == "" {
		for _, group := range subjectInfo.Groups {
			if group == "sd-admin" {
				// service directory admins can upload its own certificate info
				// (but only if that info doesn't exist yet)
				return true, nil
			}
		}
	}

	if !helpers.VerifyFingerprint(cert, fingerprint) {
		// the fingerprint does not match the one we have on record
		return false, nil
	}

	return helpers.Verify(signedData, f.rootCert, "")
}

// Integrates a record into the directory
func (f *RecordDirectory) integrate(record *eps.SignedChangeRecord) error {
	entry, ok := f.entries[record.Record.Name]
	if !ok {
		entry = eps.MakeDirectoryEntry()
		entry.Name = record.Record.Name
	}
	if err := helpers.IntegrateChangeRecord(record, entry); err != nil {
		return err
	}
	f.entries[record.Record.Name] = entry
	return nil
}

func (f *RecordDirectory) verifyHash(record *eps.SignedChangeRecord) (bool, error) {

	submittedHash := record.Hash

	err := helpers.CalculateHash(record)

	if err != nil {
		return false, err
	}

	if submittedHash != record.Hash {
		return false, nil
	}

	return true, nil
}

// calculates the chain length for a given record
func (f *RecordDirectory) chainLength(record *eps.SignedChangeRecord, visited map[string]bool) (int, error) {

	children, ok := f.recordChildren[record.Hash]

	if !ok {
		return 1, nil
	}

	max := 0
	for _, child := range children {
		if _, ok := visited[child.Hash]; ok {
			return 0, fmt.Errorf("circular relationship detected")
		} else {
			visited[child.Hash] = true
		}
		if nm, err := f.chainLength(child, visited); err != nil {
			return 0, err
		} else if nm > max {
			max = nm
		}
	}
	return 1 + max, nil
}

// picks the best record from a series of alternatives (based on chain length)
func (f *RecordDirectory) pickRecord(alternatives []*eps.SignedChangeRecord) (*eps.SignedChangeRecord, error) {
	if len(alternatives) == 1 {
		return alternatives[0], nil
	}
	maxLength := 0
	var bestAlternative *eps.SignedChangeRecord
	for _, alternative := range alternatives {
		if cl, err := f.chainLength(alternative, map[string]bool{}); err != nil {
			return nil, err
		} else if cl > maxLength {
			maxLength = cl
			bestAlternative = alternative
		}
	}
	return bestAlternative, nil

}

func (f *RecordDirectory) update() ([]*eps.SignedChangeRecord, error) {
	if entries, err := f.dataStore.Read(); err != nil {
		return nil, err
	} else {
		changeRecords := make([]*eps.SignedChangeRecord, 0, len(entries))
		for _, entry := range entries {
			switch entry.Type {
			case SignedChangeRecordEntry:
				record := &eps.SignedChangeRecord{}
				if err := json.Unmarshal(entry.Data, &record); err != nil {
					return nil, fmt.Errorf("invalid record format!")
				}
				changeRecords = append(changeRecords, record)
			default:
				return nil, fmt.Errorf("unknown entry type found...")
			}
		}

		for _, record := range changeRecords {
			f.recordsByHash[record.Hash] = record
		}

		for _, record := range changeRecords {
			var parentHash string
			// if a parent exists we set the hash to it. Records without
			// parent (root records) will be stored under the empty hash...
			if parentRecord, ok := f.recordsByHash[record.ParentHash]; ok {
				parentHash = parentRecord.Hash
			}
			children, ok := f.recordChildren[parentHash]
			if !ok {
				children = make([]*eps.SignedChangeRecord, 0)
			}
			children = append(children, record)
			f.recordChildren[parentHash] = children
		}

		records := make([]*eps.SignedChangeRecord, 0)

		currentRecords, ok := f.recordChildren[""]

		// no records present it seems
		if !ok {
			return records, nil
		}

		for {
			bestRecord, err := f.pickRecord(currentRecords)

			if err != nil {
				return nil, err
			}

			records = append(records, bestRecord)

			currentRecords, ok = f.recordChildren[bestRecord.Hash]

			if !ok {
				break
			}
		}

		for i, record := range records {
			// we verify the signature of the record (without )
			if ok, err := f.verifySignature(record, records[:i]); err != nil {
				return nil, err
			} else if !ok {
				eps.Log.Warning("signature does not match, ignoring the remainder...")
				records = records[:i]
				break
			}
		}

		// we store the ordered sequence of records
		f.orderedRecords = records

		// we regenerate the entries based on the new set of records
		f.entries = make(map[string]*eps.DirectoryEntry)
		for _, record := range records {
			f.integrate(record)
		}

		return records, nil
	}
}
