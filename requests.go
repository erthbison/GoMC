package gomc

import (
	"gomc/request"
	"reflect"
)

// Create a new Request that will be sent to a node.
//
// Id: The id of the node
// Method: The name of the method to be called on the node
// Params: The parameters to be passed to the method
func NewRequest(id int, method string, params ...any) request.Request {
	valueParams := make([]reflect.Value, len(params))
	for i, val := range params {
		valueParams[i] = reflect.ValueOf(val)
	}
	return request.Request{
		Id:     id,
		Method: method,
		Params: valueParams,
	}
}
