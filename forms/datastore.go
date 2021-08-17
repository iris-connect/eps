package forms

import (
	"github.com/kiprotect/go-helpers/forms"
)

var DatastoreForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "type",
			Validators: []forms.Validator{
				forms.IsString{},
				IsValidDatastoreType{},
			},
		},
		{
			Name: "settings",
			Validators: []forms.Validator{
				forms.IsStringMap{},
				AreValidDatastoreSettings{},
			},
		},
	},
}
