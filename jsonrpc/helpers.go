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
	"encoding/json"
	"github.com/iris-gateway/eps"
	"github.com/iris-gateway/eps/http"
	"regexp"
)

var jsonContentTypeRegexp = regexp.MustCompile("(?i)^application/json(?:;.*)?$")

// extracts the request data from
func ExtractJSONRequestData(c *http.Context) {
	eps.Log.Debugf("Extracting JSON data...")

	if !jsonContentTypeRegexp.MatchString(c.Request.Header.Get("content-type")) {
		c.JSON(400, http.H{"message": "JSON required"})
		c.Abort()
		return
	}

	var jsonData map[string]interface{}

	if err := json.NewDecoder(c.Request.Body).Decode(&jsonData); err != nil {
		c.JSON(400, http.H{"message": "invalid JSON"})
		c.Abort()
		return
	}

	if validJSON, err := JSONRPCRequestForm.Validate(jsonData); err != nil {
		// validation errors are safe to pass back to the client
		c.JSON(400, http.H{"message": "invalid JSON", "error": err})
		c.Abort()
		return
	} else {
		var requestData RequestData
		// this should never happen if the form validation is correct...
		if err := JSONRPCRequestForm.Coerce(&requestData, validJSON); err != nil {
			eps.Log.Error(err)
			c.JSON(500, http.H{"message": "internal server error"})
			c.Abort()
		}
		c.Set("requestData", &requestData)
	}

}
