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

type AreValidDatastoreSettings struct {
}

func (f AreValidDatastoreSettings) Validate(input interface{}, inputs map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("cannot validate without context")
}

func (f AreValidDatastoreSettings) ValidateWithContext(input interface{}, inputs map[string]interface{}, context map[string]interface{}) (interface{}, error) {
	definitions, ok := context["definitions"].(*eps.Definitions)
	if !ok {
		return nil, fmt.Errorf("expected a 'definitions' context")
	}
	datastoreType := inputs["type"].(string)
	// string type has been validated before
	settings := input.(map[string]interface{})
	if definition, ok := definitions.DatastoreDefinitions[datastoreType]; !ok {
		return nil, fmt.Errorf("invalid datastore type: '%s'", datastoreType)
	} else if definition.SettingsValidator == nil {
		return nil, fmt.Errorf("cannot validate settings for datastore of type '%s'", datastoreType)
	} else if validatedSettings, err := definition.SettingsValidator(settings); err != nil {
		return nil, err
	} else {
		return validatedSettings, nil
	}
}

type IsValidDatastoreType struct {
}

func (f IsValidDatastoreType) Validate(input interface{}, inputs map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("cannot validate without context")
}

func (f IsValidDatastoreType) ValidateWithContext(input interface{}, inputs map[string]interface{}, context map[string]interface{}) (interface{}, error) {
	definitions, ok := context["definitions"].(*eps.Definitions)
	if !ok {
		return nil, fmt.Errorf("expected a 'definitions' context")
	}
	// string type has been validated before
	strValue := input.(string)
	if _, ok := definitions.DatastoreDefinitions[strValue]; !ok {
		return nil, fmt.Errorf("invalid datastore type: '%s'", strValue)
	}
	return input, nil
}
