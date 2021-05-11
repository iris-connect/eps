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
	"github.com/iris-gateway/eps/helpers"
	"github.com/kiprotect/go-helpers/forms"
)

type RecordDirectorySettings struct {
	Path string `json:"path"`
}

var RecordDirectorySettingsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "path",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
	},
}

type RecordDirectory struct {
	eps.BaseDirectory
	dataStore helpers.DataStore
	Settings  RecordDirectorySettings
	Directory *Directory
}

func RecordDirectorySettingsValidator(settings map[string]interface{}) (interface{}, error) {
	if params, err := RecordDirectorySettingsForm.Validate(settings); err != nil {
		return nil, err
	} else {
		validatedSettings := &RecordDirectorySettings{}
		if err := RecordDirectorySettingsForm.Coerce(validatedSettings, params); err != nil {
			return nil, err
		}
		return validatedSettings, nil
	}
}

func MakeRecordDirectory(name string, settings interface{}) (eps.Directory, error) {
	f := &RecordDirectory{
		BaseDirectory: eps.BaseDirectory{
			Name_: name,
		},
		Settings: settings.(RecordDirectorySettings),
	}

	return f, f.load()
}

func (f *RecordDirectory) Entries(*eps.DirectoryQuery) ([]*eps.DirectoryEntry, error) {
	return nil, nil
}

func (f *RecordDirectory) OwnEntry() (*eps.DirectoryEntry, error) {
	return nil, nil
}

func (f *RecordDirectory) load() error {
	return nil
	/*
		if entries, err := p.dataStore.Read(); err != nil {
			return err
		} else {
			for _, entry := range entries {
				var data map[string]interface{}
				if err := json.Unmarshal(entry.Data, &data); err != nil {
					kodex.Log.Errorf("Error when unmarshalling entry '%s', skipping", hex.EncodeToString(entry.ID))
					continue
				}
				switch entry.Type {
				case ParametersType:
					if parameters, err := p.inMemoryStore.RestoreParameters(data); err != nil {
						return err
					} else {
						// we check if there already is a parameter set for this action and parameter
						// group. If yes, we do not overwrite it.
						if existingParameters, err := p.inMemoryStore.Parameters(parameters.Action(), parameters.ParameterGroup()); err != nil {
							return err
						} else if existingParameters != nil {
	*/
}
