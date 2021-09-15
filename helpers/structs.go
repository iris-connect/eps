package helpers

import (
	"encoding/json"
)

// convert an arbitrary structure to a string map via the JSON encoding
func ToStringMap(value interface{}) (map[string]interface{}, error) {
	var out map[string]interface{}
	if jsonData, err := json.Marshal(value); err != nil {
		return nil, err
	} else if err := json.Unmarshal(jsonData, &out); err != nil {
		return nil, err
	} else {
		return out, nil
	}
}
