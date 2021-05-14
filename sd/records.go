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
	"github.com/iris-gateway/eps/helpers"
	"net/url"
	"sort"
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
	rootCert  *x509.Certificate
	dataStore helpers.DataStore
	settings  *RecordDirectorySettings
	entries   map[string]*eps.DirectoryEntry
	records   []*eps.SignedChangeRecord
	mutex     sync.Mutex
}

func MakeRecordDirectory(settings *RecordDirectorySettings) (*RecordDirectory, error) {

	cert, err := helpers.LoadCertificate(settings.CACertificateFile, false)

	if err != nil {
		return nil, err
	}

	f := &RecordDirectory{
		rootCert:  cert,
		records:   make([]*eps.SignedChangeRecord, 0),
		entries:   make(map[string]*eps.DirectoryEntry),
		settings:  settings,
		dataStore: helpers.MakeFileDataStore(settings.DatabaseFile),
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

	// operators can edit their own channels
	if subjectInfo.Name == record.Record.Name && record.Record.Section == "channels" {
		return true, nil
	}

	for _, group := range subjectInfo.Groups {
		if group == "sd-admin" {
			// service directory admins can do everything
			return true, nil
		}
	}

	// to do: check that the signature is actually still valid

	return false, nil
}

// Appends a new record
func (f *RecordDirectory) Append(record *eps.SignedChangeRecord) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// we verify the signature of the record
	if ok, err := f.verifySignature(record); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("signature does not match")
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

	if ok, err := f.verifyHash(record, tip); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("stale append, please try again")
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
	if len(f.records) == 0 {
		return nil, nil
	}
	return f.records[len(f.records)-1], nil
}

// Returns all records since a given date
func (f *RecordDirectory) Records(since int64) ([]*eps.SignedChangeRecord, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	relevantRecords := make([]*eps.SignedChangeRecord, 0)
	for _, record := range f.records {
		if record.Position >= since {
			relevantRecords = append(relevantRecords, record)
		}
	}
	return relevantRecords, nil
}

func (f *RecordDirectory) verifySignature(record *eps.SignedChangeRecord) (bool, error) {
	signedData := &eps.SignedData{
		Data:      record,
		Signature: record.Signature,
	}

	// we temporarily remove the signature from the signed record
	signature := record.Signature
	record.Signature = nil
	defer func() { record.Signature = signature }()

	return helpers.Verify(signedData, f.rootCert, "")
}

type ByPosition struct {
	Records []*eps.SignedChangeRecord
}

func (b ByPosition) Len() int {
	return len(b.Records)
}

func (b ByPosition) Swap(i, j int) {
	b.Records[i], b.Records[j] = b.Records[j], b.Records[i]
}

func (b ByPosition) Less(i, j int) bool {

	return b.Records[i].Position < b.Records[j].Position
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

func (f *RecordDirectory) verifyHash(record, lastRecord *eps.SignedChangeRecord) (bool, error) {

	if lastRecord != nil {
		if lastRecord.Position+1 != record.Position {
			return false, nil
		}
	}

	lastHash := ""

	if lastRecord != nil {
		lastHash = lastRecord.Hash
	}

	reconstructedHash, err := helpers.CalculateHash(record, lastHash)

	if err != nil {
		return false, err
	}

	if reconstructedHash != record.Hash {
		return false, nil
	}

	return true, nil
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
				// we verify the signature of the record
				if ok, err := f.verifySignature(record); err != nil {
					return nil, err
				} else if !ok {
					eps.Log.Warning("signature does not match, ignoring...")
					continue
				}
				changeRecords = append(changeRecords, record)
			default:
				return nil, fmt.Errorf("unknown entry type found...")
			}
		}
		bp := &ByPosition{changeRecords}
		sort.Sort(bp)
		var lastRecord *eps.SignedChangeRecord

		if len(f.records) > 0 {
			lastRecord = f.records[len(f.records)-1]
		}
		// now we validate the hash chain
		for _, record := range bp.Records {

			if ok, err := f.verifyHash(record, lastRecord); err != nil {
				return nil, err
			} else if !ok {
				eps.Log.Warning("stale record found, ignoring...")
				continue
			}

			lastRecord = record
		}
		for _, record := range bp.Records {
			f.integrate(record)
		}

		allRecords := &ByPosition{append(f.records, bp.Records...)}
		sort.Sort(allRecords)
		f.records = allRecords.Records
		return bp.Records, nil
	}
}
