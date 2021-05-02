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

import ()

type DirectoryDefinition struct {
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	Maker             DirectoryMaker    `json:"-"`
	SettingsValidator SettingsValidator `json:"-"`
}

type DirectoryEntry struct {
	Name     string             `json:"name"`
	Channels []*OperatorChannel `json:"channels"`
}

type OperatorChannel struct {
	Type     string                 `json:"type"`
	Settings map[string]interface{} `json:"settings,omitempty"`
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
	Entries(*DirectoryQuery) []*DirectoryEntry
	Name() string
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
		}
		if !found {
			continue
		}
		relevantEntries = append(relevantEntries, entry)
	}
	return relevantEntries
}
