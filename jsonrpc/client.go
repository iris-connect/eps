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
	"bytes"
	"encoding/json"
	"github.com/iris-connect/eps"
	"github.com/iris-connect/eps/tls"
	"io/ioutil"
	"net/http"
)

type Client struct {
	settings *JSONRPCClientSettings
}

func MakeClient(settings *JSONRPCClientSettings) *Client {
	return &Client{
		settings: settings,
	}
}

func (c *Client) SetServerName(serverName string) {
	c.settings.ServerName = serverName
}

func (c *Client) SetEndpoint(endpoint string) {
	c.settings.Endpoint = endpoint
}

func (c *Client) Call(request *Request) (*Response, error) {
	data, err := json.Marshal(request)

	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	if c.settings.TLS != nil {
		eps.Log.Info("Using TLS")
		tlsConfig, err := tls.TLSClientConfig(c.settings.TLS, c.settings.ServerName)
		if err != nil {
			return nil, err
		}
		client.Transport = &http.Transport{
			DisableKeepAlives: true, // removing this will cause connections to pile up
			//MaxIdleConnsPerHost: 100,
			TLSClientConfig: tlsConfig,
		}
	}

	eps.Log.Infof("Generating request to endpoint %s...", c.settings.Endpoint)

	req, err := http.NewRequest("POST", c.settings.Endpoint, bytes.NewReader(data))

	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	eps.Log.Info("Done with request...")

	// to do: sanity checks...

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		return nil, err
	}

	response := &Response{}

	if err := json.Unmarshal(body, response); err != nil {
		return nil, err
	}

	return response, nil
}
