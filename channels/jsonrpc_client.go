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

package channels

import (
	"fmt"
	"github.com/iris-connect/eps"
	"github.com/iris-connect/eps/jsonrpc"
	"regexp"
)

var MethodNameRegexp = regexp.MustCompile(`(?i)^([^\.]+)\.(.*)$`)

type JSONRPCClientChannel struct {
	eps.BaseChannel
	Settings *jsonrpc.JSONRPCClientSettings
}

func JSONRPCClientSettingsValidator(settings map[string]interface{}) (interface{}, error) {
	if params, err := jsonrpc.JSONRPCClientSettingsForm.Validate(settings); err != nil {
		return nil, err
	} else {
		validatedSettings := &jsonrpc.JSONRPCClientSettings{}
		if err := jsonrpc.JSONRPCClientSettingsForm.Coerce(validatedSettings, params); err != nil {
			return nil, err
		}
		return validatedSettings, nil
	}
}

func MakeJSONRPCClientChannel(settings interface{}) (eps.Channel, error) {
	rpcSettings := settings.(jsonrpc.JSONRPCClientSettings)
	return &JSONRPCClientChannel{
		Settings: &rpcSettings,
	}, nil
}

func (c *JSONRPCClientChannel) Open() error {

	return nil
}

func (c *JSONRPCClientChannel) Close() error {
	return nil
}

func (c *JSONRPCClientChannel) DeliverRequest(request *eps.Request) (*eps.Response, error) {

	client := jsonrpc.MakeClient(c.Settings)
	jsonrpcRequest := &jsonrpc.Request{}
	jsonrpcRequest.FromEPSRequest(request)

	if groups := MethodNameRegexp.FindStringSubmatch(jsonrpcRequest.Method); groups == nil {
		return nil, fmt.Errorf("invalid method name")
	} else {
		// we remove the operator name from the method call before passing it in
		jsonrpcRequest.Method = groups[2]
	}

	jsonrpcResponse, err := client.Call(jsonrpcRequest)
	if err != nil {
		eps.Log.Error(err)
		return nil, err
	}
	return jsonrpcResponse.ToEPSResponse(), nil
}

func (c *JSONRPCClientChannel) CanDeliverTo(address *eps.Address) bool {

	if address.Operator == c.Directory().Name() {
		return true
	}

	return false
}
