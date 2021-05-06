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
	"github.com/iris-gateway/eps"
	"github.com/iris-gateway/eps/grpc"
	"github.com/kiprotect/go-helpers/forms"
)

type GRPCClientChannel struct {
	eps.BaseChannel
	Settings grpc.GRPCClientSettings
}

type GRPCClientEntrySettings struct {
	Address string `json:"address"`
}

var GRPCClientEntrySettingsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "address",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
	},
}

func getEntrySettings(settings map[string]interface{}) (*GRPCClientEntrySettings, error) {
	if params, err := GRPCClientEntrySettingsForm.Validate(settings); err != nil {
		return nil, err
	} else {
		validatedSettings := &GRPCClientEntrySettings{}
		if err := GRPCClientEntrySettingsForm.Coerce(validatedSettings, params); err != nil {
			return nil, err
		}
		return validatedSettings, nil
	}
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
	return nil
}

func (c *GRPCClientChannel) Close() error {
	return nil
}

func (c *GRPCClientChannel) DeliverRequest(request *eps.Request) (*eps.Response, error) {

	address, err := eps.GetAddress(request.ID)

	if err != nil {
		return nil, err
	}

	entry, err := c.DirectoryEntry(address, "grpc_server")

	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, fmt.Errorf("cannot deliver request")
	}

	if len(entry.Channels) == 0 {
		return nil, fmt.Errorf("cannot find channel")
	}

	settings, err := getEntrySettings(entry.Channels[0].Settings)

	if err != nil {
		return nil, err
	}

	if client, err := grpc.MakeClient(&c.Settings); err != nil {
		return nil, err
	} else if err := client.Connect(settings.Address, entry.Name); err != nil {
		return nil, err
	} else {
		return client.SendRequest(request)
	}
}

func (c *GRPCClientChannel) DeliverResponse(response *eps.Response) error {
	return nil
}

func (c *GRPCClientChannel) CanDeliverTo(address *eps.Address) bool {

	// we check if the requested service offers a gRPC server
	if entry, err := c.DirectoryEntry(address, "grpc_server"); entry != nil {
		return true
	} else if err != nil {
		// we log this error
		eps.Log.Error(err)
	}

	return false
}
