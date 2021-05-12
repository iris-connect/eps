package forms

import (
	"github.com/kiprotect/go-helpers/forms"
)

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
