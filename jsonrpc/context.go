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

type Context struct {
	Request *Request
}

func (c *Context) Result(data interface{}) *Response {
	return &Response{
		ID:      &c.Request.ID,
		Result:  data,
		JSONRPC: "2.0",
	}
}

func (c *Context) Error(code int, message string, data interface{}) *Response {
	return &Response{
		Error: &Error{
			Code:    code,
			Message: message,
			Data:    data,
		},
		JSONRPC: "2.0",
		ID:      &c.Request.ID,
	}
}

func (c *Context) MethodNotFound() *Response {
	return c.Error(-32601, "method not found", nil)
}

func (c *Context) InvalidParams(err error) *Response {
	return c.Error(-32602, "invalid params", err)
}

func (c *Context) InternalError() *Response {
	return c.Error(-32603, "intenal error", nil)
}
