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
	"github.com/iris-gateway/eps/grpc"
)

type GRPCServerChannel struct {
	eps.BaseChannel
	server   *grpc.Server
	Settings grpc.GRPCServerSettings
}

func GRPCServerSettingsValidator(settings map[string]interface{}) (interface{}, error) {
	if params, err := grpc.GRPCServerSettingsForm.Validate(settings); err != nil {
		return nil, err
	} else {
		validatedSettings := &grpc.GRPCServerSettings{}
		if err := grpc.GRPCServerSettingsForm.Coerce(validatedSettings, params); err != nil {
			return nil, err
		}
		return validatedSettings, nil
	}
}

func MakeGRPCServerChannel(settings interface{}) (eps.Channel, error) {
	return &GRPCServerChannel{
		Settings: settings.(grpc.GRPCServerSettings),
	}, nil
}

func (c *GRPCServerChannel) HandleRequest(request *eps.Request, clientInfo *eps.ClientInfo) (*eps.Response, error) {
	return c.MessageBroker().DeliverRequest(request, clientInfo)
}

func (c *GRPCServerChannel) Open() error {
	var err error
	if c.server, err = grpc.MakeServer(&c.Settings, c, c.Directory()); err != nil {
		return err
	}
	return c.server.Start()
}

func (c *GRPCServerChannel) Close() error {
	return nil
}

func (c *GRPCServerChannel) DeliverRequest(request *eps.Request) (*eps.Response, error) {
	return c.server.DeliverRequest(request)
}

func (c *GRPCServerChannel) CanDeliverTo(address *eps.Address) bool {
	return c.server.CanDeliverTo(address)
}
