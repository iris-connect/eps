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
	"encoding/hex"
	"github.com/iris-gateway/eps"
	"github.com/iris-gateway/eps/forms"
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
	if err := forms.DirectoryEntryForm.Coerce(entry, config); err != nil {
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

func CalculateHash(record *eps.SignedChangeRecord) error {

	// we always reset the hash before calculating the new one
	record.Hash = ""

	hash, err := StructuredHash(record.Record)

	if err != nil {
		return err
	}

	record.Hash = hex.EncodeToString(hash[:])

	return nil

}
