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
	"github.com/iris-connect/eps"
)

var Channels = eps.ChannelDefinitions{
	"stdout": eps.ChannelDefinition{
		Name:              "Stdout Channel",
		Description:       "Prints messages to stdout (just for testing and debugging)",
		Maker:             MakeStdoutChannel,
		SettingsValidator: StdoutSettingsValidator,
	},
	"jsonrpc_client": eps.ChannelDefinition{
		Name:              "JSONRPC Client Channel",
		Description:       "Creates outgoing JSONRPC connections to deliver and receive messages",
		Maker:             MakeJSONRPCClientChannel,
		SettingsValidator: JSONRPCClientSettingsValidator,
	},
	"grpc_client": eps.ChannelDefinition{
		Name:              "gRPC Client Channel",
		Description:       "Creates outgoing gRPC connections to deliver and receive messages",
		Maker:             MakeGRPCClientChannel,
		SettingsValidator: GRPCClientSettingsValidator,
	},
	"jsonrpc_server": eps.ChannelDefinition{
		Name:              "JSONRPC Server Channel",
		Description:       "Accepts incoming JSONRPC connections to deliver and receive messages",
		Maker:             MakeJSONRPCServerChannel,
		SettingsValidator: JSONRPCServerSettingsValidator,
	},
	"grpc_server": eps.ChannelDefinition{
		Name:              "gRPC Server Channel",
		Description:       "Accepts incoming gRPC connections to deliver and receive messages",
		Maker:             MakeGRPCServerChannel,
		SettingsValidator: GRPCServerSettingsValidator,
	},
}
