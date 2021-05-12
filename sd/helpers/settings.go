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
	"github.com/iris-gateway/eps/sd"
	"github.com/kiprotect/go-helpers/settings"
	"os"
	"strings"
)

var EnvSettingsName = "SD_SETTINGS"

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

func Settings(settingsPaths []string) (*sd.Settings, error) {
	if rawSettings, err := settings.MakeSettings(settingsPaths); err != nil {
		return nil, err
	} else if params, err := sd.SettingsForm.Validate(rawSettings.Values); err != nil {
		return nil, err
	} else {
		settings := &sd.Settings{}
		if err := sd.SettingsForm.Coerce(settings, params); err != nil {
			// this should not happen if the forms are correct...
			return nil, err
		}
		// settings are valid
		return settings, nil
	}
}
