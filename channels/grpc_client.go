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

type GRPCClientChannel struct {
	eps.BaseChannel
	client   *grpc.Client
	Settings grpc.GRPCClientSettings
}

func GRPCClientSettingsValidator(settings map[string]interface{}) (interface{}, error) {
	if params, err := grpc.GRPCClientSettingsForm.Validate(settings); err != nil {
		return nil, err
	} else {
		validatedSettings := &grpc.GRPCClientSettings{}
		if err := grpc.GRPCClientSettingsForm.Coerce(validatedSettings, params); err != nil {
			return nil, err
		}
		return validatedSettings, nil
	}
}

func MakeGRPCClientChannel(settings interface{}) (eps.Channel, error) {
	return &GRPCClientChannel{
		Settings: settings.(grpc.GRPCClientSettings),
	}, nil
}

func (c *GRPCClientChannel) Open() error {
	var err error
	if c.client, err = grpc.MakeClient(&c.Settings); err != nil {
		return err
	}
	return nil
}

func (c *GRPCClientChannel) Close() error {
	return nil
}

func (c *GRPCClientChannel) DeliverRequest(request *eps.Request) (*eps.Response, error) {
	return nil, nil
}

func (c *GRPCClientChannel) DeliverResponse(response *eps.Response) error {
	return nil
}

func (c *GRPCClientChannel) CanDeliverTo(address *eps.Address) bool {
	return false
}
