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
	"github.com/iris-gateway/eps/tls"
)

// Settings for the JSON-RPC server
type JSONRPCClientSettings struct {
	TLS      *tls.TLSSettings `json:"tls"`
	Endpoint string           `json:"endpoint"`
	Enabled  bool             `json:"enabled"`
}

// Settings for the JSON-RPC server
type JSONRPCServerSettings struct {
	TLS         *tls.TLSSettings `json:"tls"`
	BindAddress string           `json:"bind_address"`
	Enabled     bool             `json:"enabled"`
}
