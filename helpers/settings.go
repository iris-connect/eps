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

// This code was extracted from the Kodex CE project
// (https://github.com/kiprotect/kodex), the author has
// the copyright to the code so no attribution is necessary.

package helpers

import (
	"github.com/iris-connect/eps"
	epsForms "github.com/iris-connect/eps/forms"
	"github.com/kiprotect/go-helpers/forms"
	"github.com/kiprotect/go-helpers/settings"
	"os"
	"strings"
)

var EnvSettingsName = "EPS_SETTINGS"

func SettingsPaths() []string {
	envValue := os.Getenv(EnvSettingsName)
	if envValue == "" {
		return []string{}
	}
	values := strings.Split(envValue, ":")
	sanitizedValues := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		sanitizedValues = append(sanitizedValues, value)
	}
	return sanitizedValues
}

func Settings(settingsPaths []string, definitions *eps.Definitions) (*eps.Settings, error) {
	if rawSettings, err := settings.MakeSettings(settingsPaths); err != nil {
		return nil, err
	} else if params, err := epsForms.SettingsForm.ValidateWithContext(rawSettings.Values, map[string]interface{}{"definitions": definitions}); err != nil {
		return nil, err
	} else {
		settings := &eps.Settings{
			Definitions: definitions,
		}
		if err := forms.Coerce(settings, params); err != nil {
			// this should not happen if the forms are correct...
			return nil, err
		}
		// settings are valid
		return settings, nil
	}
}
