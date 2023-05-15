package request

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
