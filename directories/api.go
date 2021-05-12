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
	"github.com/iris-gateway/eps"
	"github.com/kiprotect/go-helpers/forms"
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
	},
}

type APIDirectorySettings struct {
	Endpoints []string `json:"endpoints"`
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
	Settings APIDirectorySettings
	Cache    *DirectoryCache
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
	d := &APIDirectory{
		BaseDirectory: eps.BaseDirectory{
			Name_: name,
		},
		Settings: settings.(APIDirectorySettings),
	}

	return d, nil
}

func (f *APIDirectory) Entries(query *eps.DirectoryQuery) ([]*eps.DirectoryEntry, error) {
	return nil, nil
}

func (f *APIDirectory) OwnEntry() (*eps.DirectoryEntry, error) {
	return nil, nil
}
