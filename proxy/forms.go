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
