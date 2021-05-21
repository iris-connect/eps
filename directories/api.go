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

package directories

import (
	"crypto/x509"
	"fmt"
	"github.com/iris-gateway/eps"
	epsForms "github.com/iris-gateway/eps/forms"
	"github.com/iris-gateway/eps/helpers"
	"github.com/iris-gateway/eps/jsonrpc"
	"github.com/kiprotect/go-helpers/forms"
	"sync"
	"time"
)

var APIDirectorySettingsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "endpoints",
			Validators: []forms.Validator{
				forms.IsStringList{},
			},
		},
		{
			Name: "server_names",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsStringList{},
			},
		},
		{
			Name: "jsonrpc_client",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &jsonrpc.JSONRPCClientSettingsForm,
				},
			},
		},
		{
			Name: "ca_certificate_files",
			Validators: []forms.Validator{
				forms.IsList{
					Validators: []forms.Validator{
						forms.IsString{},
					},
				},
			},
		},
	},
}

type APIDirectorySettings struct {
	Endpoints          []string                       `json:"endpoints"`
	ServerNames        []string                       `json:"server_names"`
	JSONRPCClient      *jsonrpc.JSONRPCClientSettings `json:"jsonrpc_client"`
	CACertificateFiles []string                       `json:"ca_certificate_files"`
}

type CacheEntry struct {
	Entry     *eps.DirectoryEntry
	FetchedAt time.Time
}

type DirectoryCache struct {
	Entries []*CacheEntry
}

type APIDirectory struct {
	eps.BaseDirectory
	settings      APIDirectorySettings
	jsonrpcClient *jsonrpc.Client
	rootCerts     []*x509.Certificate
	entries       map[string]*eps.DirectoryEntry
	records       []*eps.SignedChangeRecord
	mutex         sync.Mutex
}

func APIDirectorySettingsValidator(settings map[string]interface{}) (interface{}, error) {
	if params, err := APIDirectorySettingsForm.Validate(settings); err != nil {
		return nil, err
	} else {
		validatedSettings := &APIDirectorySettings{}
		if err := APIDirectorySettingsForm.Coerce(validatedSettings, params); err != nil {
			return nil, err
		}
		return validatedSettings, nil
	}
}

func MakeAPIDirectory(name string, settings interface{}) (eps.Directory, error) {
	apiSettings := settings.(APIDirectorySettings)

	rootCerts := make([]*x509.Certificate, 0)

	for _, certificateFile := range apiSettings.CACertificateFiles {
		cert, err := helpers.LoadCertificate(certificateFile, false)

		if err != nil {
			return nil, err
		}

		rootCerts = append(rootCerts, cert)

	}

	d := &APIDirectory{
		BaseDirectory: eps.BaseDirectory{
			Name_: name,
		},
		jsonrpcClient: jsonrpc.MakeClient(apiSettings.JSONRPCClient),
		entries:       make(map[string]*eps.DirectoryEntry),
		records:       []*eps.SignedChangeRecord{},
		rootCerts:     rootCerts,
		settings:      apiSettings,
	}

	return d, d.update()
}

var UpdateForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "records",
			Validators: []forms.Validator{
				forms.IsList{
					Validators: []forms.Validator{
						forms.IsStringMap{
							Form: &epsForms.SignedChangeRecordForm,
						},
					},
				},
			},
		},
	},
}

type UpdateRecords struct {
	Records []*eps.SignedChangeRecord `json:"records"`
}

func (f *APIDirectory) Entries(query *eps.DirectoryQuery) ([]*eps.DirectoryEntry, error) {

	f.mutex.Lock()
	defer f.mutex.Unlock()

	if err := f.update(); err != nil {
		return nil, err
	}
	entries := make([]*eps.DirectoryEntry, len(f.entries))
	i := 0
	for _, entry := range f.entries {
		entries[i] = entry
		i++
	}
	return eps.FilterDirectoryEntriesByQuery(entries, query), nil
}

func (f *APIDirectory) EntryFor(name string) (*eps.DirectoryEntry, error) {
	// locking is done by Entries method
	if entries, err := f.Entries(&eps.DirectoryQuery{Operator: name}); err != nil {
		return nil, err
	} else if len(entries) == 0 {
		return nil, fmt.Errorf("no entry for %s", name)
	} else {
		return entries[0], nil
	}
}

func (f *APIDirectory) OwnEntry() (*eps.DirectoryEntry, error) {
	// locking is done by Entries method
	return f.EntryFor(f.Name())
}

func (f *APIDirectory) Tip() (*eps.SignedChangeRecord, error) {

	// to do: ensure there's always one server name and endpoint
	f.jsonrpcClient.SetServerName(f.settings.ServerNames[0])
	f.jsonrpcClient.SetEndpoint(f.settings.Endpoints[0])

	request := jsonrpc.MakeRequest("getTip", "", map[string]interface{}{})

	if result, err := f.jsonrpcClient.Call(request); err != nil {
		return nil, err
	} else {
		if result.Error != nil {
			return nil, fmt.Errorf(result.Error.Message)
		}

		if result.Result == nil {
			return nil, nil
		}

		if mapResult, ok := result.Result.(map[string]interface{}); !ok {
			return nil, fmt.Errorf("expected a map")
		} else if params, err := epsForms.SignedChangeRecordForm.Validate(mapResult); err != nil {
			return nil, err
		} else {
			signedChangeRecord := &eps.SignedChangeRecord{}
			if err := epsForms.SignedChangeRecordForm.Coerce(signedChangeRecord, params); err != nil {
				return nil, err
			} else {
				return signedChangeRecord, nil
			}
		}
	}

	return nil, nil
}

func (f *APIDirectory) Submit(signedChangeRecords []*eps.SignedChangeRecord) error {
	// to do: ensure there's always one server name and endpoint
	f.jsonrpcClient.SetServerName(f.settings.ServerNames[0])
	f.jsonrpcClient.SetEndpoint(f.settings.Endpoints[0])

	// we tell the internal proxy about an incoming connection
	request := jsonrpc.MakeRequest("submitRecords", "", map[string]interface{}{"records": signedChangeRecords})

	if result, err := f.jsonrpcClient.Call(request); err != nil {
		return err
	} else {
		if result.Error != nil {
			eps.Log.Error(result.Error)
			return fmt.Errorf(result.Error.Message)
		}
		return nil
	}

}

func (f *APIDirectory) integrate(records []*eps.SignedChangeRecord) error {
	for _, record := range records {
		entry, ok := f.entries[record.Record.Name]
		if !ok {
			entry = eps.MakeDirectoryEntry()
			entry.Name = record.Record.Name
		}
		if err := helpers.IntegrateChangeRecord(record, entry); err != nil {
			return err
		}
		f.entries[record.Record.Name] = entry
	}
	return nil
}

// Updates the service directory with change records from the remote API
func (f *APIDirectory) update() error {
	// to do: ensure there's always one server name and endpoint
	f.jsonrpcClient.SetServerName(f.settings.ServerNames[0])
	f.jsonrpcClient.SetEndpoint(f.settings.Endpoints[0])

	var tipHash string

	if len(f.records) > 0 {
		tipHash = f.records[len(f.records)-1].Hash
	}

	// we tell the internal proxy about an incoming connection
	request := jsonrpc.MakeRequest("getRecords", "", map[string]interface{}{"after": tipHash})

	if result, err := f.jsonrpcClient.Call(request); err != nil {
		return err
	} else {

		if result.Error != nil {
			return fmt.Errorf(result.Error.Message)
		}

		if result.Result == nil {
			return nil
		}

		config := map[string]interface{}{
			"records": result.Result,
		}

		if params, err := UpdateForm.Validate(config); err != nil {
			return err
		} else {
			updateRecords := &UpdateRecords{}
			if err := UpdateForm.Coerce(updateRecords, params); err != nil {
				return err
			} else {
				records := updateRecords.Records

				var fullRecords []*eps.SignedChangeRecord
				var resetEntries bool

				if len(records) > 0 && records[0].ParentHash != tipHash {
					if records[0].ParentHash != "" {
						return fmt.Errorf("expected a new root record but got one with parent hash '%s'", records[0].ParentHash)
					}
					// seems the directory changed, we make sure the new one is actually newer than the current one
					if len(f.records) > 0 && records[0].Record.CreatedAt.Time.Before(f.records[0].Record.CreatedAt.Time) {
						return fmt.Errorf("server tried to provide an outdated service directory")
					} else {
						eps.Log.Warning("Service directory root changed!")
						// we reset the entries
						fullRecords = records
						resetEntries = true
					}
				} else {
					fullRecords = append(f.records, records...)
				}

				// we verify all records before we integate them
				for i, record := range fullRecords {
					if ok, err := helpers.VerifyRecord(record, fullRecords[:i], f.rootCerts); err != nil {
						return err
					} else if !ok {
						return fmt.Errorf("invalid record found")
					}
				}

				f.records = fullRecords
				if resetEntries {
					f.entries = make(map[string]*eps.DirectoryEntry)
				}

				// we integrate the new records
				return f.integrate(records)
			}
		}
	}
}
