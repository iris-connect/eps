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

package eps

import (
	"github.com/kiprotect/go-helpers/forms"
)

type TLSSettings struct {
	CACertificateFile string `json:"ca_certificate_file"`
	CertificateFile   string `json:"certificate_file"`
	KeyFile           string `json:"key_file"`
}

// Settings for the gRPC server
type GRPCServerSettings struct {
	TLS         *TLSSettings `json:"tls"`
	BindAddress string       `json:"bind_address"`
	Enabled     bool         `json:"enabled"`
}

// Settings for the gRPC client
type GRPCClientSettings struct {
	TLS     *TLSSettings `json:"tls"`
	Enabled bool         `json:"enabled"`
}

// Settings for the JSON-RPC server
type JSONRPCServerSettings struct {
	TLS         *TLSSettings `json:"tls"`
	BindAddress string       `json:"bind_address"`
	Enabled     bool         `json:"enabled"`
}

type Settings struct {
	GRPCClient    *GRPCClientSettings    `json:"grpc_client"`
	GRPCServer    *GRPCServerSettings    `json:"grpc_server"`
	JSONRPCServer *JSONRPCServerSettings `json:"jsonrpc_server"`
}

var TLSSettingsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "ca_certificate_file",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name: "certificate_file",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name: "key_file",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
	},
}

var GRPCClientSettingsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "tls",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &TLSSettingsForm,
				},
			},
		},
	},
}

var GRPCServerSettingsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "bind_address",
			Validators: []forms.Validator{
				forms.IsOptional{Default: "localhost:4444"},
				forms.IsString{},
			},
		},
		{
			Name: "tls",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &TLSSettingsForm,
				},
			},
		},
	},
}

var JSONRPCServerSettingsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "tls",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &TLSSettingsForm,
				},
			},
		},
	},
}

var SettingsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "grpc_client",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsStringMap{
					Form: &GRPCClientSettingsForm,
				},
			},
		},
		{
			Name: "grpc_server",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsStringMap{
					Form: &GRPCServerSettingsForm,
				},
			},
		},
		{
			Name: "jsonrpc_server",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsStringMap{
					Form: &JSONRPCServerSettingsForm,
				},
			},
		},
	},
}
