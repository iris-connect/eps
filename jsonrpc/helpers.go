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

package jsonrpc

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/iris-gateway/eps"
	"github.com/iris-gateway/eps/helpers"
	"github.com/iris-gateway/eps/http"
	"regexp"
	"strings"
)

var jsonContentTypeRegexp = regexp.MustCompile("(?i)^application/json(?:;.*)?$")

// extracts the request data from
func ExtractJSONRequest(c *http.Context) {
	eps.Log.Debugf("Extracting JSON data...")

	invalidJSONResponse := Response{JSONRPC: "2.0,", Error: &Error{Code: -32700, Message: "JSON required"}}
	invalidRequestResponse := func(err error) *Response {
		return &Response{JSONRPC: "2.0,", Error: &Error{Code: -32600, Message: "invalid request", Data: err}}
	}
	serverErrorResponse := Response{JSONRPC: "2.0,", Error: &Error{Code: -32603, Message: "internal server error"}}

	if !jsonContentTypeRegexp.MatchString(c.Request.Header.Get("content-type")) {
		c.JSON(400, invalidJSONResponse)
		return
	}

	var jsonData map[string]interface{}

	if err := json.NewDecoder(c.Request.Body).Decode(&jsonData); err != nil {
		c.JSON(400, invalidJSONResponse)
		return
	}

	if validJSON, err := JSONRPCRequestForm.Validate(jsonData); err != nil {
		// validation errors are safe to pass back to the client
		c.JSON(400, invalidRequestResponse(err))
		return
	} else {
		var request Request

		id, ok := validJSON["id"]

		// if no ID is contained we generate a random UUID
		if !ok {
			if randomID, err := helpers.RandomBytes(16); err != nil {
				c.JSON(500, serverErrorResponse)
				return
			} else {
				validJSON["id"] = hex.EncodeToString(randomID)
			}
		} else {
			switch v := id.(type) {
			case int64:
				// we convert numbers to strings
				validJSON["id"] = fmt.Sprintf("n:%d", v)
			case string:
				if matches := idNRegexp.FindStringSubmatch(v); matches != nil {
					// we need to escape the string IDs that match our custom format...
					validJSON["id"] = fmt.Sprintf("%s:%s", strings.Repeat("n", 2*len(matches[1])), matches[2])
				}
			}
		}

		// this should never happen if the form validation is correct...
		if err := JSONRPCRequestForm.Coerce(&request, validJSON); err != nil {
			eps.Log.Error(err)
			c.JSON(500, serverErrorResponse)
			return
		}

		c.Set("request", &request)
	}

}
