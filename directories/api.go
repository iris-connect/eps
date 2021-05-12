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

package directories

import (
	"fmt"
	"github.com/iris-gateway/eps"
	epsForms "github.com/iris-gateway/eps/forms"
	"github.com/iris-gateway/eps/jsonrpc"
	"github.com/kiprotect/go-helpers/forms"
	"time"
)

var APIDirectorySettingsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "endpoints",
			Validators: []forms.Validator{
				forms.IsStringList{},
			},
		},
		{
			Name: "server_names",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsStringList{},
			},
		},
		{
			Name: "jsonrpc_client",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &jsonrpc.JSONRPCClientSettingsForm,
				},
			},
		},
	},
}

type APIDirectorySettings struct {
	Endpoints     []string                       `json:"endpoints"`
	ServerNames   []string                       `json:"server_names"`
	JSONRPCClient *jsonrpc.JSONRPCClientSettings `json:"jsonrpc_client"`
}

type CacheEntry struct {
	Entry     *eps.DirectoryEntry
	FetchedAt time.Time
}

type DirectoryCache struct {
	Entries []*CacheEntry
}

type APIDirectory struct {
	eps.BaseDirectory
	settings      APIDirectorySettings
	jsonrpcClient *jsonrpc.Client
	Cache         *DirectoryCache
}

func APIDirectorySettingsValidator(settings map[string]interface{}) (interface{}, error) {
	if params, err := APIDirectorySettingsForm.Validate(settings); err != nil {
		return nil, err
	} else {
		validatedSettings := &APIDirectorySettings{}
		if err := APIDirectorySettingsForm.Coerce(validatedSettings, params); err != nil {
			return nil, err
		}
		return validatedSettings, nil
	}
}

func MakeAPIDirectory(name string, settings interface{}) (eps.Directory, error) {
	apiSettings := settings.(APIDirectorySettings)
	d := &APIDirectory{
		BaseDirectory: eps.BaseDirectory{
			Name_: name,
		},
		jsonrpcClient: jsonrpc.MakeClient(apiSettings.JSONRPCClient),
		settings:      apiSettings,
	}

	return d, nil
}

func (f *APIDirectory) Entries(query *eps.DirectoryQuery) ([]*eps.DirectoryEntry, error) {
	return nil, nil
}

func (f *APIDirectory) OwnEntry() (*eps.DirectoryEntry, error) {
	return nil, nil
}

func (f *APIDirectory) Tip() (*eps.SignedChangeRecord, error) {

	// to do: ensure there's always one server name and endpoint
	f.jsonrpcClient.SetServerName(f.settings.ServerNames[0])
	f.jsonrpcClient.SetEndpoint(f.settings.Endpoints[0])

	// we tell the internal proxy about an incoming connection
	request := jsonrpc.MakeRequest("getTip", "", map[string]interface{}{})

	if result, err := f.jsonrpcClient.Call(request); err != nil {
		return nil, err
	} else {
		if result.Error != nil {
			return nil, fmt.Errorf(result.Error.Message)
		}

		if result.Result == nil {
			return nil, nil
		}

		if mapResult, ok := result.Result.(map[string]interface{}); !ok {
			return nil, fmt.Errorf("expected a map")
		} else if params, err := epsForms.SignedChangeRecordForm.Validate(mapResult); err != nil {
			return nil, err
		} else {
			signedChangeRecord := &eps.SignedChangeRecord{}
			if err := epsForms.SignedChangeRecordForm.Coerce(signedChangeRecord, params); err != nil {
				return nil, err
			} else {
				return signedChangeRecord, nil
			}
		}
	}

	return nil, nil
}

func (f *APIDirectory) Submit(signedChangeRecord *eps.SignedChangeRecord) error {
	// to do: ensure there's always one server name and endpoint
	f.jsonrpcClient.SetServerName(f.settings.ServerNames[0])
	f.jsonrpcClient.SetEndpoint(f.settings.Endpoints[0])

	// we tell the internal proxy about an incoming connection
	request := jsonrpc.MakeRequest("submitChangeRecord", "", map[string]interface{}{"record": signedChangeRecord})

	if result, err := f.jsonrpcClient.Call(request); err != nil {
		return err
	} else {
		if result.Error != nil {
			eps.Log.Error(result.Error)
			return fmt.Errorf(result.Error.Message)
		}
		return nil
	}

}
