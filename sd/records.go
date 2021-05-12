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
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/iris-gateway/eps"
	"github.com/iris-gateway/eps/helpers"
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
	eps.BaseDirectory
	rootCert  *x509.Certificate
	dataStore helpers.DataStore
	settings  *RecordDirectorySettings
	entries   []*eps.DirectoryEntry
	records   []*eps.SignedChangeRecord
	mutex     sync.Mutex
}

func MakeRecordDirectory(name string, settings *RecordDirectorySettings) (*RecordDirectory, error) {

	cert, err := eps.LoadCertificate(settings.CACertificateFile, false)

	if err != nil {
		return nil, err
	}

	f := &RecordDirectory{
		BaseDirectory: eps.BaseDirectory{
			Name_: name,
		},
		rootCert: cert,
		records:  make([]*eps.SignedChangeRecord, 0),
		entries:  make([]*eps.DirectoryEntry, 0),
		settings: settings,
	}

	return f, f.update()
}

func (f *RecordDirectory) Entries(*eps.DirectoryQuery) ([]*eps.DirectoryEntry, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return nil, nil
}

// Appends a new record
func (f *RecordDirectory) Append(record *eps.SignedChangeRecord) error {
	return nil
}

func (f *RecordDirectory) Update() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.update()
}

// Returns the latest record
func (f *RecordDirectory) Tip() (*eps.SignedChangeRecord, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return nil, nil
}

// Returns all records since a given date
func (f *RecordDirectory) Records(since int) ([]*eps.SignedChangeRecord, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return nil, nil
}

func (f *RecordDirectory) verifySignature(record *eps.SignedChangeRecord) (bool, error) {
	signedData := &eps.SignedData{
		Data:      record.Record,
		Signature: record.Signature,
	}

	return eps.Verify(signedData, f.rootCert, record.Record.Name)
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

// Integrates a new record into the directory
func (f *RecordDirectory) integrate(record *eps.SignedChangeRecord) error {
	return nil
}

/*
 */
func (f *RecordDirectory) update() error {
	if entries, err := f.dataStore.Read(); err != nil {
		return err
	} else {
		changeRecords := make([]*eps.SignedChangeRecord, 0, len(entries))
		for _, entry := range entries {
			switch entry.Type {
			case SignedChangeRecordEntry:
				record := &eps.SignedChangeRecord{}
				if err := json.Unmarshal(entry.Data, &record); err != nil {
					return fmt.Errorf("invalid record format!")
				}
				// we verify the signature of the record
				if ok, err := f.verifySignature(record); err != nil {
					return err
				} else if !ok {
					return fmt.Errorf("signature does not match")
				}
				changeRecords = append(changeRecords, record)
			default:
				return fmt.Errorf("unknown entry type found...")
			}
		}
		bp := &ByPosition{changeRecords}
		sort.Sort(bp)
		var lastRecord *eps.SignedChangeRecord
		// now we validate the hash chain
		for _, record := range bp.Records {
			if lastRecord != nil {
				if lastRecord.Position+1 != record.Position {
					return fmt.Errorf("missing record")
				}
			}

			rawData, err := json.Marshal(record.Record)

			if err != nil {
				return err
			}

			hash := sha256.Sum256(rawData)
			lastRecordHash, err := hex.DecodeString(lastRecord.Hash)

			if err != nil {
				return err
			}

			// we construct new hash data from the hash of the last record, the hash of the current
			// record and the position in the hash chain
			fullHashData := append(append(lastRecordHash, hash[:]...), []byte(fmt.Sprintf("%d", record.Position))...)

			fullHash := sha256.Sum256(fullHashData)

			if hex.EncodeToString(fullHash[:]) != record.Hash {
				return fmt.Errorf("chain hashes do not match")
			}
		}
		for _, record := range bp.Records {
			f.integrate(record)
		}
	}
	return nil
}
