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

// The JSON directory loads the service directory from a single JSON file.
// This is just for testing, for production use please use the "file" directory
// which provides support for signed service directory records.

package directories

import (
	"encoding/json"
	"fmt"
	"github.com/iris-gateway/eps"
	epsForms "github.com/iris-gateway/eps/forms"
	"github.com/kiprotect/go-helpers/forms"
	"io/ioutil"
	"os"
)

var JSONDirectorySettingsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "path",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
	},
}

var JSONDirectoryForm = forms.Form{
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

type JSONDirectorySettings struct {
	Path string `json:"path"`
}

type Directory struct {
	Entries eps.DirectoryEntries `json:"entries"`
}

type JSONDirectory struct {
	eps.BaseDirectory
	Settings  JSONDirectorySettings
	Directory *Directory
}

func JSONDirectorySettingsValidator(settings map[string]interface{}) (interface{}, error) {
	if params, err := JSONDirectorySettingsForm.Validate(settings); err != nil {
		return nil, err
	} else {
		validatedSettings := &JSONDirectorySettings{}
		if err := JSONDirectorySettingsForm.Coerce(validatedSettings, params); err != nil {
			return nil, err
		}
		return validatedSettings, nil
	}
}

func MakeJSONDirectory(name string, settings interface{}) (eps.Directory, error) {
	d := &JSONDirectory{
		BaseDirectory: eps.BaseDirectory{
			Name_: name,
		},
		Settings: settings.(JSONDirectorySettings),
	}

	return d, d.load()
}

func (f *JSONDirectory) load() error {
	if directory, err := LoadJSONDirectory(f.Settings.Path); err != nil {
		return err
	} else {
		eps.Log.Debugf("Loaded %d directory entries...", len(directory.Entries))
		f.Directory = directory
		return nil
	}
}

func (f *JSONDirectory) Entries(query *eps.DirectoryQuery) ([]*eps.DirectoryEntry, error) {
	return eps.FilterDirectoryEntriesByQuery(f.Directory.Entries, query), nil
}

func (f *JSONDirectory) OwnEntry() (*eps.DirectoryEntry, error) {
	if entries, err := f.Entries(&eps.DirectoryQuery{Operator: f.Name()}); err != nil {
		return nil, err
	} else if len(entries) == 0 {
		return nil, fmt.Errorf("no entry for myself")
	} else {
		return entries[0], nil
	}
}

func LoadJSONDirectory(path string) (*Directory, error) {
	if file, err := os.Open(path); err != nil {
		return nil, err
	} else {
		defer file.Close()
		if data, err := ioutil.ReadAll(file); err != nil {
			return nil, err
		} else {
			rawDirectory := map[string]interface{}{}
			if err := json.Unmarshal(data, &rawDirectory); err != nil {
				return nil, err
			} else if params, err := JSONDirectoryForm.Validate(rawDirectory); err != nil {
				return nil, err
			} else {
				directory := &Directory{}
				if err := forms.Coerce(directory, params); err != nil {
					// this should not happen if the forms are correct...
					return nil, err
				}
				// directory is validate
				return directory, nil
			}
		}
	}
}
