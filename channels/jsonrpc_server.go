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

type JSONRPCServerChannel struct {
	eps.BaseChannel
	Settings *jsonrpc.JSONRPCServerSettings
	Server   *jsonrpc.JSONRPCServer
}

func JSONRPCServerSettingsValidator(settings map[string]interface{}) (interface{}, error) {
	if params, err := jsonrpc.JSONRPCServerSettingsForm.Validate(settings); err != nil {
		return nil, err
	} else {
		validatedSettings := &jsonrpc.JSONRPCServerSettings{}
		if err := jsonrpc.JSONRPCServerSettingsForm.Coerce(validatedSettings, params); err != nil {
			return nil, err
		}
		return validatedSettings, nil
	}
}

func MakeJSONRPCServerChannel(broker eps.MessageBroker, settings interface{}) (eps.Channel, error) {
	rpcSettings := settings.(jsonrpc.JSONRPCServerSettings)

	s := &JSONRPCServerChannel{
		BaseChannel: eps.BaseChannel{Broker: broker},
		Settings:    &rpcSettings,
	}

	methods := []*jsonrpc.Method{
		{
			Name:    "submit",
			Handler: s.submitData,
		},
	}

	if server, err := jsonrpc.MakeJSONRPCServer(&rpcSettings, methods); err != nil {
		return nil, err
	} else {
		s.Server = server
		return s, nil
	}
}

func (c *JSONRPCServerChannel) submitData(context *jsonrpc.Context) *jsonrpc.Response {
	return context.Error(200, "test", nil)
}

func (c *JSONRPCServerChannel) Open() error {
	return c.Server.Start()
}

func (c *JSONRPCServerChannel) Close() error {
	return c.Server.Stop()
}

func (c *JSONRPCServerChannel) Deliver(message *eps.Message) (*eps.Message, error) {
	return nil, nil
}

func (c *JSONRPCServerChannel) CanDeliver(message *eps.Message) bool {
	return false
}
func (c *JSONRPCServerChannel) CanHandle(message *eps.Message) bool {
	return false
}
