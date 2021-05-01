package forms

import (
	"fmt"
	"github.com/iris-gateway/eps"
)

type AreValidChannelSettings struct {
}

func (f AreValidChannelSettings) Validate(input interface{}, inputs map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("cannot validate without context")
}

func (f AreValidChannelSettings) ValidateWithContext(input interface{}, inputs map[string]interface{}, context map[string]interface{}) (interface{}, error) {
	definitions, ok := context["definitions"].(*eps.Definitions)
	if !ok {
		return nil, fmt.Errorf("expected a 'definitions' context")
	}
	channelType := inputs["type"].(string)
	// string type has been validated before
	settings := input.(map[string]interface{})
	if definition, ok := definitions.ChannelDefinitions[channelType]; !ok {
		return nil, fmt.Errorf("invalid channel type: '%s'", channelType)
	} else if definition.SettingsValidator == nil {
		return nil, fmt.Errorf("cannot validate settings for channel of type '%s'", channelType)
	} else if validatedSettings, err := definition.SettingsValidator(settings); err != nil {
		return nil, err
	} else {
		return validatedSettings, nil
	}
}

type IsValidChannelType struct {
}

func (f IsValidChannelType) Validate(input interface{}, inputs map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("cannot validate without context")
}

func (f IsValidChannelType) ValidateWithContext(input interface{}, inputs map[string]interface{}, context map[string]interface{}) (interface{}, error) {
	definitions, ok := context["definitions"].(*eps.Definitions)
	if !ok {
		return nil, fmt.Errorf("expected a 'definitions' context")
	}
	// string type has been validated before
	strValue := input.(string)
	if _, ok := definitions.ChannelDefinitions[strValue]; !ok {
		return nil, fmt.Errorf("invalid channel type: '%s'", strValue)
	}
	return input, nil
}
