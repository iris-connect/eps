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
	"github.com/kiprotect/go-helpers/forms"
)

var SettingsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "public",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsStringMap{
					Form: &PublicSettingsForm,
				},
			},
		},
		{
			Name: "private",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsStringMap{
					Form: &PrivateSettingsForm,
				},
			},
		},
	},
}
var PrivateSettingsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "bind_address",
			Validators: []forms.Validator{
				forms.IsOptional{Default: "localhost:443"},
				forms.IsString{},
			},
		},
		{
			Name: "eps_endpoint",
			Validators: []forms.Validator{
				forms.IsOptional{Default: "localhost:5555"},
				forms.IsString{},
			},
		},
		{
			Name: "jsonrpc_server",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &jsonrpc.JSONRPCServerSettingsForm,
				},
			},
		},
	},
}

var PublicSettingsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "bind_address",
			Validators: []forms.Validator{
				forms.IsOptional{Default: "localhost:443"},
				forms.IsString{},
			},
		},
		{
			Name: "eps_endpoint",
			Validators: []forms.Validator{
				forms.IsOptional{Default: "localhost:5555"},
				forms.IsString{},
			},
		},
		{
			Name: "jsonrpc_server",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &jsonrpc.JSONRPCServerSettingsForm,
				},
			},
		},
	},
}
