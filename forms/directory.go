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
)

var OperatorChannelForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "type",
			Validators: []forms.Validator{
				forms.IsString{},
				// we do not validate the channel type because it can contain
				// channel types that are not implemented by the local server
				// which does not mean that they can't exist though...
			},
		},
		{
			Name: "settings",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsStringMap{},
			},
		},
	},
}

var OperatorCertificateForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "serial_number",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name: "key_usage",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
	},
}

var ServiceValidatorForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "type",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name: "parameters",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsStringMap{},
			},
		},
	},
}

var ServiceParameterForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "name",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name: "validators",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsList{
					Validators: []forms.Validator{
						forms.IsStringMap{
							Form: &ServiceValidatorForm,
						},
					},
				},
			},
		},
	},
}

var OperatorServiceForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "name",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name: "authorized_clients",
			Validators: []forms.Validator{
				forms.IsOptional{Default: []interface{}{}},
				forms.IsList{
					Validators: []forms.Validator{
						forms.IsString{},
					},
				},
			},
		},
		{
			Name: "parameters",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsList{
					Validators: []forms.Validator{
						forms.IsStringMap{
							Form: &ServiceParameterForm,
						},
					},
				},
			},
		},
	},
}

var DirectoryEntryForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "name",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name: "channels",
			Validators: []forms.Validator{
				forms.IsList{
					Validators: []forms.Validator{
						forms.IsStringMap{
							Form: &OperatorChannelForm,
						},
					},
				},
			},
		},
		{
			Name: "services",
			Validators: []forms.Validator{
				forms.IsList{
					Validators: []forms.Validator{
						forms.IsStringMap{
							Form: &OperatorServiceForm,
						},
					},
				},
			},
		},
		{
			Name: "certificates",
			Validators: []forms.Validator{
				forms.IsList{
					Validators: []forms.Validator{
						forms.IsStringMap{
							Form: &OperatorCertificateForm,
						},
					},
				},
			},
		},
	},
}

var SignedChangeRecordForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "position",
			Validators: []forms.Validator{
				forms.IsInteger{
					HasMin: true,
					Min:    0,
				},
			},
		},
		{
			Name: "hash",
			Validators: []forms.Validator{
				forms.IsHex{
					Strict:    true, // we don't allow any '-' characters
					MinLength: 32,   // this is the binary length
					MaxLength: 32,   // this is the binary length
				},
			},
		},
		{
			Name: "signature",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &SignatureForm,
				},
			},
		},
		{
			Name: "record",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &ChangeRecordForm,
				},
			},
		},
	},
}

var ChangeRecordForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "name",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name: "created_at",
			Validators: []forms.Validator{
				forms.IsString{},
				forms.IsTime{
					Format: "rfc3339",
				},
			},
		},
		{
			Name: "section",
			Validators: []forms.Validator{
				forms.IsString{},
				forms.IsIn{
					Choices: []interface{}{"channels", "certificates", "services"},
				},
			},
		},
		{
			Name: "data",
			Validators: []forms.Validator{
				forms.Switch{
					Key: "section",
					Cases: map[string][]forms.Validator{
						"channels": []forms.Validator{
							forms.IsList{
								Validators: []forms.Validator{
									forms.IsStringMap{
										Form: &OperatorChannelForm,
									},
								},
							},
						},
						"services": []forms.Validator{
							forms.IsList{
								Validators: []forms.Validator{
									forms.IsStringMap{
										Form: &OperatorServiceForm,
									},
								},
							},
						},
						"certificates": []forms.Validator{
							forms.IsList{
								Validators: []forms.Validator{
									forms.IsStringMap{
										Form: &OperatorCertificateForm,
									},
								},
							},
						},
					},
				},
			},
		},
	},
}
