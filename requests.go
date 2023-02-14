package gomc

import "reflect"

// Represent a function call or request to a node
// Id: The id of the node
// Method: The name of the method to be called on the node
// Params: The parameters to be passed to the method
type Request struct {
	Id     int
	Method string
	Params []reflect.Value
}

// Create a new Func object.
// Id: The id of the node
// Method: The name of the method to be called on the node
// Params: The parameters to be passed to the method
func NewRequest(id int, method string, params ...any) Request {
	test := make([]reflect.Value, len(params))
	for i, val := range params {
		test[i] = reflect.ValueOf(val)
	}
	return Request{
		Id:     id,
		Method: method,
		Params: test,
	}
}
