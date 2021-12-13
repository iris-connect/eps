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

package fixtures

import (
	"fmt"
	"github.com/iris-connect/eps"
	"github.com/iris-connect/eps/definitions"
	"github.com/iris-connect/eps/helpers"
	"path"
)

type Settings struct {
	Paths           []string
	EnvSettingsName string
}

func (c Settings) Setup(fixtures map[string]interface{}) (interface{}, error) {
	// we set the loglevel to 'debug' so we can see which settings files are being loaded
	var defs *eps.Definitions
	var paths []string
	var ok bool
	if defs, ok = fixtures["definitions"].(*eps.Definitions); !ok {
		defs = &definitions.Default
	}

	if c.EnvSettingsName == "" {
		c.EnvSettingsName = "EPS_SETTINGS"
	}

	settingsPaths, fs, err := helpers.SettingsPaths(c.EnvSettingsName)

	if err != nil {
		return nil, err
	}

	if c.Paths != nil {
		if len(settingsPaths) != 1 {
			return nil, fmt.Errorf("expected a single settings path prefix")
		}
		fullPaths := []string{}
		for _, pth := range c.Paths {
			fullPaths = append(fullPaths, path.Join(append(settingsPaths, pth)...))
		}
		paths = fullPaths
	} else {
		paths = settingsPaths
	}
	eps.Log.SetLevel(eps.DebugLogLevel)
	return helpers.Settings(paths, fs, defs)
}

func (c Settings) Teardown(fixture interface{}) error {
	return nil
}
