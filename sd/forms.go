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

package sd

import (
	epsForms "github.com/iris-connect/eps/forms"
	"github.com/iris-connect/eps/jsonrpc"
	"github.com/kiprotect/go-helpers/forms"
)

var RecordDirectorySettingsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "datastore",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &epsForms.DatastoreForm,
				},
			},
		},
		{
			Name: "metrics",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsStringMap{
					Form: &epsForms.MetricsSettingsForm,
				},
			},
		},
		{
			Name: "ca_certificate_files",
			Validators: []forms.Validator{
				forms.IsList{
					Validators: []forms.Validator{
						forms.IsString{},
					},
				},
			},
		},
		{
			Name: "ca_intermediate_certificate_files",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsList{
					Validators: []forms.Validator{
						forms.IsString{},
					},
				},
			},
		},
	},
}

var SettingsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "directory",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &RecordDirectorySettingsForm,
				},
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
