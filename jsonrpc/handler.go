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
	"fmt"
	"github.com/iris-gateway/eps"
	"github.com/kiprotect/go-helpers/forms"
	"reflect"
)

type Method struct {
	Form    *forms.Form
	Handler interface{}
}

type Coercer func(interface{}) error

// returns a new struct that we can coerce the valid form parameters into
func handlerStruct(handler interface{}) (interface{}, error) {
	value := reflect.ValueOf(handler)
	if value.Kind() != reflect.Func {
		return nil, fmt.Errorf("not a function")
	}

	funcType := value.Type()

	if funcType.NumIn() != 2 {
		return nil, fmt.Errorf("expected a function with 2 arguments")
	}

	if funcType.NumOut() != 1 {
		return nil, fmt.Errorf("expected a function with 1 return value")
	}

	returnType := funcType.Out(0)

	if !returnType.AssignableTo(reflect.TypeOf(&Response{})) {
		return nil, fmt.Errorf("return value should be a response")
	}

	contextType := funcType.In(0)

	if !contextType.AssignableTo(reflect.TypeOf(&Context{})) {
		return nil, fmt.Errorf("first argument should accept a context")
	}

	structType := funcType.In(1)

	if structType.Kind() != reflect.Ptr || structType.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("second argument should be a struct pointer")
	}

	// we create a new struct and return it
	return reflect.New(structType.Elem()).Interface(), nil
}

// calls the handler with the validated and coerced form parameters
func callHandler(context *Context, handler, params interface{}) (*Response, error) {
	value := reflect.ValueOf(handler)

	if value.Kind() != reflect.Func {
		return nil, fmt.Errorf("not a function")
	}

	paramsValue := reflect.ValueOf(params)
	contextValue := reflect.ValueOf(context)

	responseValue := value.Call([]reflect.Value{contextValue, paramsValue})

	return responseValue[0].Interface().(*Response), nil

}

// A more advanced handler that performs form-based input sanization and coerces
// the result into a user-provided data structure via reflection.
func MethodsHandler(methods map[string]*Method) (Handler, error) {

	// we check that all provided methods have the correct type
	for key, method := range methods {
		if _, err := handlerStruct(method.Handler); err != nil {
			return nil, err
		}
		if method.Form == nil {
			return nil, fmt.Errorf("form for method %s missing", key)
		}
	}

	return func(context *Context) *Response {
		if method, ok := methods[context.Request.Method]; !ok {
			return context.MethodNotFound()
		} else {
			if params, err := method.Form.ValidateWithContext(context.Request.Params, map[string]interface{}{"context": context}); err != nil {
				return context.InvalidParams(err)
			} else {
				if paramsStruct, err := handlerStruct(method.Handler); err != nil {
					// this should never happen...
					eps.Log.Error(err)
					return context.InternalError()
				} else if err := method.Form.Coerce(paramsStruct, params); err != nil {
					// this shouldn't happen either...
					eps.Log.Error(err)
					return context.InternalError()
				} else {
					if response, err := callHandler(context, method.Handler, paramsStruct); err != nil {
						// and neither should this...
						eps.Log.Error(err)
						return context.InternalError()
					} else {
						return response
					}
				}
			}
		}
	}, nil
}
