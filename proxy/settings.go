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

package proxy

import (
	"github.com/iris-gateway/eps/jsonrpc"
)

type Settings struct {
	Private *PrivateServerSettings `json:"private"`
	Public  *PublicServerSettings  `json:"public"`
}

type PublicServerSettings struct {
	TLSBindAddress      string                         `json:"tls_bind_address"`
	InternalBindAddress string                         `json:"internal_bind_address"`
	JSONRPCClient       *jsonrpc.JSONRPCClientSettings `json:"jsonrpc_client"`
	JSONRPCServer       *jsonrpc.JSONRPCServerSettings `json:"jsonrpc_server`
}

type PrivateServerSettings struct {
	EPSEndpoint   string                         `json:"eps_endpoint "`
	JSONRPCClient *jsonrpc.JSONRPCClientSettings `json:"jsonrpc_client"`
	JSONRPCServer *jsonrpc.JSONRPCServerSettings `json:"jsonrpc_server`
}
