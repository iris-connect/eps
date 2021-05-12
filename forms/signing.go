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

package forms

import (
	"github.com/kiprotect/go-helpers/forms"
	"regexp"
)

var SignedDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "signature",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &SignatureForm,
				},
			},
		},
		{
			Name:       "Data",
			Validators: []forms.Validator{}, // can be anything really...
		},
	},
}

var SignatureForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "R",
			Validators: []forms.Validator{
				forms.IsString{},
				forms.MatchesRegex{
					Regex: regexp.MustCompile(`^\d{10,100}$`),
				},
			},
		},
		{
			Name: "S",
			Validators: []forms.Validator{
				forms.IsString{},
				forms.MatchesRegex{
					Regex: regexp.MustCompile(`^\d{10,100}$`),
				},
			},
		},
		{
			Name: "C",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
	},
}

var SigningSettingsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "ca_certificate_file",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsString{},
			},
		},
		{
			Name: "certificate_file",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsString{},
			},
		},
		{
			Name: "key_file",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsString{},
			},
		},
		{
			Name: "name",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsString{},
			},
		},
	},
}
