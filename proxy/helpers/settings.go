package helpers

import (
	"github.com/iris-gateway/eps/proxy"
	"github.com/kiprotect/go-helpers/settings"
	"os"
	"strings"
)

var EnvSettingsName = "PROXY_SETTINGS"

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

func Settings(settingsPaths []string) (*proxy.Settings, error) {
	if rawSettings, err := settings.MakeSettings(settingsPaths); err != nil {
		return nil, err
	} else if params, err := proxy.SettingsForm.Validate(rawSettings.Values); err != nil {
		return nil, err
	} else {
		settings := &proxy.Settings{}
		if err := proxy.SettingsForm.Coerce(settings, params); err != nil {
			// this should not happen if the forms are correct...
			return nil, err
		}
		// settings are valid
		return settings, nil
	}
}
