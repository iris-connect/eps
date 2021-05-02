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
	"encoding/json"
	"github.com/iris-gateway/eps"
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

type DirectoryEntry struct {
	Name     string             `json:"string"`
	Channels []*OperatorChannel `json:"channels"`
}

type OperatorChannel struct {
	Type     string                 `json:"type"`
	Settings map[string]interface{} `json:"settings,omitempty"`
}

var JSONOperatorChannelForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "type",
			Validators: []forms.Validator{
				forms.IsString{},
				// we do not validate the channel type because it can contain
				// channel types that are not implemented by the local server
				// which does not mean that they can't exist though...
			},
		},
		{
			Name: "settings",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsStringMap{},
			},
		},
	},
}

var JSONDirectoryEntryForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "name",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name: "channels",
			Validators: []forms.Validator{
				forms.IsList{
					Validators: []forms.Validator{
						forms.IsStringMap{
							Form: &JSONOperatorChannelForm,
						},
					},
				},
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
							Form: &JSONDirectoryEntryForm,
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
		if err := JSONDirectoryForm.Coerce(validatedSettings, params); err != nil {
			return nil, err
		}
		return validatedSettings, nil
	}
}

func MakeJSONDirectory(settings interface{}) (eps.Directory, error) {
	d := &JSONDirectory{
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

func (f *JSONDirectory) Entries() []*eps.DirectoryEntry {
	return nil
}

type Directory struct {
	Entries eps.DirectoryEntries `json:"entries"`
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
				// directory is valid
				return directory, nil
			}
		}
	}
}
