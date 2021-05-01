package forms

import (
	"fmt"
	"github.com/iris-gateway/eps"
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
