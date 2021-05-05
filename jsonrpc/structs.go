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
	"github.com/iris-gateway/eps"
)

// we always convert incoming IDs to strings
type Request struct {
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
	ID      string                 `json:"id"`
}

func (r *Request) FromEPSRequest(request *eps.Request) {
	r.JSONRPC = "2.0"
	r.Method = request.Method
	r.ID = request.ID
	r.Params = request.Params
}

type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

func (r *Response) ToEPSResponse() *eps.Response {

	strId, ok := r.ID.(string)

	if !ok {
		eps.Log.Warningf("Warning, non-string response ID found: %v", r.ID)
	}

	response := &eps.Response{
		ID: &strId,
	}

	if r.Result != nil {
		mapResult, ok := r.Result.(map[string]interface{})
		if !ok {
			mapResult = map[string]interface{}{"_": r.Result}
		}
		response.Result = mapResult
	}

	if r.Error != nil {
		error := &eps.Error{
			Code:    r.Error.Code,
			Message: r.Error.Message,
		}
		if r.Error.Data != nil {
			mapErrorData, ok := r.Error.Data.(map[string]interface{})
			if !ok {
				mapErrorData = map[string]interface{}{"_": r.Error.Data}
			}
			error.Data = mapErrorData
		}
		response.Error = error
	}
	return response
}

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
