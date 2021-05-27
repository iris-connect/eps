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
	"github.com/iris-connect/eps/tls"
	"github.com/kiprotect/go-helpers/forms"
)

var JSONRPCRequestForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "jsonrpc",
			Validators: []forms.Validator{
				forms.IsString{},
				forms.IsIn{
					// we only support JSONRPC-2.0 right now
					Choices: []interface{}{"2.0"},
				},
			},
		},
		{
			Name: "method",
			Validators: []forms.Validator{
				forms.IsString{
					MinLength: 1,
					MaxLength: 100,
				},
			},
		},
		{
			Name: "params",
			Validators: []forms.Validator{
				// we only support string-map style parameter passing
				forms.IsStringMap{},
			},
		},
		{
			Name: "id",
			Validators: []forms.Validator{
				// it may be omitted (then we generate one)
				forms.IsOptional{},
				// either a string or an integer
				forms.Or{
					Options: [][]forms.Validator{
						{
							// we support strings
							forms.IsString{
								MinLength: 1,
								MaxLength: 100,
							},
						},
						{
							// we also support integers (but only 32 bit)
							forms.IsInteger{
								HasMin: true,
								HasMax: true,
								Min:    -2147483648,
								Max:    2147483647,
							},
						},
					},
				},
			},
		},
	},
}

var CorsSettingsForm = forms.Form{
	ErrorMsg: "invalid data encountered in the CORS settings form",
	Fields: []forms.Field{
		{
			Name: "allowed_hosts",
			Validators: []forms.Validator{
				forms.IsOptional{Default: []string{}},
				forms.IsStringList{},
			},
		},
		{
			Name: "allowed_headers",
			Validators: []forms.Validator{
				forms.IsOptional{Default: []string{}},
				forms.IsStringList{},
			},
		},
		{
			Name: "allowed_methods",
			Validators: []forms.Validator{
				forms.IsOptional{Default: []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"}},
				forms.IsList{
					Validators: []forms.Validator{
						forms.IsIn{
							Choices: []interface{}{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
						},
					},
				},
				forms.IsStringList{},
			},
		},
	},
}

var JSONRPCServerSettingsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "cors",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsStringMap{
					Form: &CorsSettingsForm,
				},
			},
		},
		{
			Name: "bind_address",
			Validators: []forms.Validator{
				forms.IsString{}, // to do: add URL validation
			},
		},
		{
			Name: "tls",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsStringMap{
					Form: &tls.TLSSettingsForm,
				},
			},
		},
	},
}

var JSONRPCClientSettingsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "endpoint",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsString{}, // to do: add URL validation
			},
		},
		{
			Name: "local",
			Validators: []forms.Validator{
				forms.IsOptional{Default: true},
				forms.IsBoolean{},
			},
		},
		{
			Name: "tls",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsStringMap{
					Form: &tls.TLSSettingsForm,
				},
			},
		},
	},
}
