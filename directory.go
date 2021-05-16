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

package eps

import (
	"time"
)

type DirectoryDefinition struct {
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	Maker             DirectoryMaker    `json:"-"`
	SettingsValidator SettingsValidator `json:"-"`
}

func MakeDirectoryEntry() *DirectoryEntry {
	return &DirectoryEntry{
		Groups:       []string{},
		Channels:     []*OperatorChannel{},
		Services:     []*OperatorService{},
		Certificates: []*OperatorCertificate{},
		Settings:     []*OperatorSettings{},
		Records:      []*SignedChangeRecord{},
	}
}

type DirectoryEntry struct {
	Name         string                 `json:"name"`
	Groups       []string               `json:"groups"`
	Channels     []*OperatorChannel     `json:"channels"`
	Services     []*OperatorService     `json:"services"`
	Certificates []*OperatorCertificate `json:"certificates"`
	Settings     []*OperatorSettings    `json:"settings"`
	Preferences  []*OperatorPreferences `json:"preferences"`
	Records      []*SignedChangeRecord  `json:"records"`
}

// preferences may be set by the corresponding operator itself
type OperatorPreferences struct {
	Operator    string                 `json:"operator"`
	Service     string                 `json:"service"`
	Environment string                 `json:"environment"`
	Preferences map[string]interface{} `json:"preferences"`
}

// settings may only be set by a directory admin
type OperatorSettings struct {
	Operator    string                 `json:"operator"`
	Service     string                 `json:"service"`
	Environment string                 `json:"environment"`
	Settings    map[string]interface{} `json:"settings"`
}

type OperatorChannel struct {
	Type     string                 `json:"type"`
	Settings map[string]interface{} `json:"settings"`
}

type OperatorCertificate struct {
	Fingerprint string `json:"fingerprint"`
	KeyUsage    string `json:"key_usage"`
}

type OperatorService struct {
	Name        string           `json:"name"`
	Permissions []*Permission    `json:"permissions"`
	Methods     []*ServiceMethod `json:"methods"`
}

type ServiceMethod struct {
	Name        string              `json:"name"`
	Permissions []*Permission       `json:"permissions"`
	Parameters  []*ServiceParameter `json:"parameters"`
}

type Permission struct {
	Group  string   `json:"group"`
	Rights []string `json:"rights"`
}

type ServiceParameter struct {
	Name       string              `json:"name"`
	Validators []*ServiceValidator `json:"validators"`
}

type ServiceValidator struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
}

type SignedChangeRecord struct {
	ParentHash string        `json:"parent_hash"`
	Hash       string        `json:"hash"`
	Signature  *Signature    `json:"signature"`
	Record     *ChangeRecord `json:"record"`
}

type Signature struct {
	R           string `json:"r"`
	S           string `json:"s"`
	Certificate string `json:"c"`
}

type SignedData struct {
	Signature *Signature  `json:"signature"`
	Data      interface{} `json:"data"`
}

// describes a change in a specific section of the service directory
type ChangeRecord struct {
	Name      string       `json:"name"`
	Section   string       `json:"section"`
	Data      interface{}  `json:"data"`
	CreatedAt HashableTime `json:"created_at"`
}

type HashableTime struct {
	time.Time
}

func (h HashableTime) HashValue() interface{} {
	return h.Time.Format(time.RFC3339)
}

func (d *DirectoryEntry) Channel(channelType string) *OperatorChannel {
	for _, channel := range d.Channels {
		if channel.Type == channelType {
			return channel
		}
	}
	return nil
}

type DirectoryQuery struct {
	Operator string
	Channels []string
}

type DirectoryEntries []*DirectoryEntry
type DirectoryDefinitions map[string]DirectoryDefinition
type DirectoryMaker func(name string, settings interface{}) (Directory, error)

// A directory can deliver and accept message
type Directory interface {
	Entries(*DirectoryQuery) ([]*DirectoryEntry, error)
	EntryFor(string) (*DirectoryEntry, error)
	OwnEntry() (*DirectoryEntry, error)
	Name() string
}

type WritableDirectory interface {
	Directory
	// required for submitting change records
	Tip() (*SignedChangeRecord, error)
	Submit(*SignedChangeRecord) error
}

type BaseDirectory struct {
	Name_ string
}

func (b *BaseDirectory) Name() string {
	return b.Name_
}

// helper function that can be used by directory implementations that
// have a list of local directory entries
func FilterDirectoryEntriesByQuery(entries []*DirectoryEntry, query *DirectoryQuery) []*DirectoryEntry {
	relevantEntries := make([]*DirectoryEntry, 0)
	for _, entry := range entries {
		// we filter the entries by the specified operator name
		if query.Operator != "" && entry.Name != query.Operator {
			continue
		}
		// we filter the entries by the specified channel types
		found := false
		if query.Channels != nil {
			for _, queryChannel := range query.Channels {
				for _, entryChannel := range entry.Channels {
					if entryChannel.Type == queryChannel {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
		} else {
			found = true
		}
		if !found {
			continue
		}
		relevantEntries = append(relevantEntries, entry)
	}
	return relevantEntries
}
