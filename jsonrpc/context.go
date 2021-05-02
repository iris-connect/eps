package jsonrpc

type Context struct {
	Request *Request
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
