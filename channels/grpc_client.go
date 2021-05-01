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
	Settings *grpc.GRPCClientSettings
}

func GRPCClientSettingsValidator(definitions *eps.Definitions, settings map[string]interface{}) (interface{}, error) {
	return settings, nil
}

func MakeGRPCClientChannel(definitions *eps.Definitions, settings interface{}) (eps.Channel, error) {
	return &GRPCClientChannel{
		Settings: settings.(*grpc.GRPCClientSettings),
	}, nil
}

func (c *GRPCClientChannel) Open() error {
	return nil
}

func (c *GRPCClientChannel) Close() error {
	return nil
}

func (c *GRPCClientChannel) Deliver(message *eps.Message) (*eps.Message, error) {
	return nil, nil
}

func (c *GRPCClientChannel) CanDeliver(message *eps.Message) bool {
	return false
}
func (c *GRPCClientChannel) CanHandle(message *eps.Message) bool {
	return false
}
