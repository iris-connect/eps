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

package forms

import (
	"fmt"
	"github.com/iris-connect/eps"
)

type AreValidDirectorySettings struct {
}

func (f AreValidDirectorySettings) Validate(input interface{}, inputs map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("cannot validate without context")
}

func (f AreValidDirectorySettings) ValidateWithContext(input interface{}, inputs map[string]interface{}, context map[string]interface{}) (interface{}, error) {
	definitions, ok := context["definitions"].(*eps.Definitions)
	if !ok {
		return nil, fmt.Errorf("expected a 'definitions' context")
	}
	directoryType := inputs["type"].(string)
	// string type has been validated before
	settings := input.(map[string]interface{})
	if definition, ok := definitions.DirectoryDefinitions[directoryType]; !ok {
		return nil, fmt.Errorf("invalid directory type: '%s'", directoryType)
	} else if definition.SettingsValidator == nil {
		return nil, fmt.Errorf("cannot validate settings for directory of type '%s'", directoryType)
	} else if validatedSettings, err := definition.SettingsValidator(settings); err != nil {
		return nil, err
	} else {
		return validatedSettings, nil
	}
}

type IsValidDirectoryType struct {
}

func (f IsValidDirectoryType) Validate(input interface{}, inputs map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("cannot validate without context")
}

func (f IsValidDirectoryType) ValidateWithContext(input interface{}, inputs map[string]interface{}, context map[string]interface{}) (interface{}, error) {
	definitions, ok := context["definitions"].(*eps.Definitions)
	if !ok {
		return nil, fmt.Errorf("expected a 'definitions' context")
	}
	// string type has been validated before
	strValue := input.(string)
	if _, ok := definitions.DirectoryDefinitions[strValue]; !ok {
		return nil, fmt.Errorf("invalid directory type: '%s'", strValue)
	}
	return input, nil
}
