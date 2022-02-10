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

package eps

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// this variable gets updated using the build process
var Version = "development"

var MethodNameRegexp = regexp.MustCompile(`(?i)^(.*)\.(.*)$`)
var IDAddressRegexp = regexp.MustCompile(`(?i)^(.*)\.([^\(\.]+)\(([^\)]+)\)$`)

type Address struct {
	Operator string `json:"operator"`
	Method   string `json:"method"`
	ID       string `json:"id"`
}

type Request struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
	ID     string                 `json:"id"`
}

type ClientInfo struct {
	Name  string          `json:"name"`
	Entry *DirectoryEntry `json:"entry"`
}

type ConnectionRequest struct {
	Endpoint string      `json:"endpoint"`
	Channel  string      `json:"channel"`
	Token    []byte      `json:"token"`
	Client   *ClientInfo `json:"client"`
}

// for inclusion in protobuf... A bit dirty as we use JSON here, but it works...
func (c *ClientInfo) AsStruct() (map[string]interface{}, error) {
	if data, err := json.Marshal(c); err != nil {
		return nil, fmt.Errorf("error marshalling as JSON: %w", err)
	} else {
		var mapStruct map[string]interface{}
		if err := json.Unmarshal(data, &mapStruct); err != nil {
			return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
		} else {
			return mapStruct, nil
		}
	}
}

func GetAddress(id string) (*Address, error) {
	if groups := IDAddressRegexp.FindStringSubmatch(id); groups == nil {
		return nil, fmt.Errorf("invalid ID format")
	} else {
		return &Address{
			Operator: groups[1],
			Method:   groups[2],
			ID:       groups[3],
		}, nil
	}
}

type Response struct {
	Result map[string]interface{} `json:"result,omitempty"`
	Error  *Error                 `json:"error,omitempty"`
	ID     *string                `json:"id"`
}

type Error struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

func ChannelError(id *string, message string, data map[string]interface{}) *Response {
	return &Response{
		ID: id,
		Error: &Error{
			Code:    500,
			Message: message,
			Data:    data,
		},
	}
}

func PermissionDenied(id *string, message string, data map[string]interface{}) *Response {
	return &Response{
		ID: id,
		Error: &Error{
			Code:    403,
			Message: message,
			Data:    data,
		},
	}
}
