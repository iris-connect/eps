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
)

var FileDirectoryForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "path",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
	},
}

type FileDirectorySettings struct {
	Path string `json:"path"`
}

type FileDirectory struct {
	eps.BaseDirectory
	Settings FileDirectorySettings
}

func FileDirectorySettingsValidator(settings map[string]interface{}) (interface{}, error) {
	if params, err := FileDirectoryForm.Validate(settings); err != nil {
		return nil, err
	} else {
		validatedSettings := &FileDirectorySettings{}
		if err := FileDirectoryForm.Coerce(validatedSettings, params); err != nil {
			return nil, err
		}
		return validatedSettings, nil
	}
}

func MakeFileDirectory(settings interface{}) (eps.Directory, error) {
	return &FileDirectory{
		Settings: settings.(FileDirectorySettings),
	}, nil
}
