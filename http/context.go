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

package http

import (
	"encoding/json"
	"github.com/iris-gateway/eps"
	"net/http"
)

type Context struct {
	Writer         http.ResponseWriter
	Request        *http.Request
	Route          *Route
	RouteParams    []string
	currentHandler int
	Aborted        bool
	HeaderWritten  bool
	values         map[string]interface{}
}

func MakeContext(writer http.ResponseWriter, request *http.Request) *Context {
	return &Context{
		Writer:  writer,
		Request: request,
		values:  make(map[string]interface{}),
	}
}

func (c *Context) Set(key string, value interface{}) {
	c.values[key] = value
}

func (c *Context) Get(key string) interface{} {
	v, _ := c.values[key]
	// we simply return nil if the value isn't set
	return v
}

func (c *Context) Abort() {
	c.Aborted = true
}

func (c *Context) AbortWithStatus(status int) {
	c.Writer.WriteHeader(status)
	c.Abort()
}

func (c *Context) JSON(status int, data interface{}) {

	if c.HeaderWritten {
		// the header was already written, we ignore this...
		eps.Log.Error("Header was already written")
		return
	}

	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	bytes, err := json.Marshal(data)

	if err != nil {
		// the data cannot be serialized to JSON, we log an error and return
		// a 500 response...
		eps.Log.Error(err)
		bytes, _ = json.Marshal(H{"message": "internal server error"})
		status = 500
	}

	c.Writer.WriteHeader(status)
	c.HeaderWritten = true

	c.Writer.Write(bytes)

	c.Abort()

}
