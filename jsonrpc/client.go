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
	"io/ioutil"
	"net/http"
)

type Client struct {
	endpoint string
}

func MakeClient(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
	}
}

func (c *Client) Call(request *Request) (*Response, error) {
	data, err := json.Marshal(request)

	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", c.endpoint, bytes.NewReader(data))

	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	// to do: sanity checks...

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	response := &Response{}

	if err := json.Unmarshal(body, response); err != nil {
		return nil, err
	}

	return response, nil
}
