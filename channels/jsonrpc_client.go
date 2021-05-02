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
	"github.com/iris-gateway/eps"
	"github.com/iris-gateway/eps/jsonrpc"
)

type JSONRPCClientChannel struct {
	eps.BaseChannel
	Settings jsonrpc.JSONRPCClientSettings
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
	return &JSONRPCClientChannel{
		Settings: settings.(jsonrpc.JSONRPCClientSettings),
	}, nil
}

func (c *JSONRPCClientChannel) Open() error {

	return nil
}

func (c *JSONRPCClientChannel) Close() error {
	return nil
}

func (c *JSONRPCClientChannel) DeliverRequest(request *eps.Request) (*eps.Response, error) {
	client := jsonrpc.MakeClient(c.Settings.Endpoint)
	eps.Log.Info("Calling!")
	jsonrpcRequest := &jsonrpc.Request{}
	jsonrpcRequest.FromEPSRequest(request)
	jsonrpcResponse, err := client.Call(jsonrpcRequest)
	if err != nil {
		return nil, err
	}
	return jsonrpcResponse.ToEPSResponse(), nil
}

func (c *JSONRPCClientChannel) DeliverResponse(response *eps.Response) error {
	return nil
}

func (c *JSONRPCClientChannel) CanDeliverTo(address *eps.Address) bool {

	if address.Operator == c.Directory().Name() {
		return true
	}

	return false
}
