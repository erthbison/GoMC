package event

import (
	"fmt"
	"reflect"
)

// An event representing some request made to a node.
type FunctionEvent struct {

	// The index of the request. Is unique among the FunctionEvents.
	index  int
	target int
	method string
	params []reflect.Value

	id EventId
}

// Create a new FunctionEvent
//
// i is the index of the FunctionEvent among the provided requests.
// target is the id of the node that will receive the request.
// method is a string with the name of the method that will be called on the node.
// params is the parameters that will be passed to the method.
func NewFunctionEvent(i int, target int, method string, params ...reflect.Value) FunctionEvent {
	return FunctionEvent{
		index:  i,
		target: target,
		method: method,
		params: params,

		id: EventId(fmt.Sprint("Function", i)),
	}
}

// An id that identifies the event.
// Two events that provided the same input state results in the same output state should have the same id
//
// New event implementations should include a identifier of the event type to prevent accidental collisions with other implementations
func (fe FunctionEvent) Id() EventId {
	return fe.id
}

func (fe FunctionEvent) String() string {
	return fmt.Sprintf("{Function %v. Target: %v}", fe.index, fe.target)
}

// A method executing the event.
// 
// Use reflection to call the specified method on the nodes with the provided parameters.
// 
// The event will be executed on a separate goroutine.
// It should signal on the channel if it is clear for the simulator to proceed to processing of the state and the next event.
// Panics raised while executing the event is recovered by the simulator and returned as errors
func (fe FunctionEvent) Execute(node any, nextEvt chan error) {
	method := reflect.ValueOf(node).MethodByName(fe.method)
	method.Call(fe.params)
	nextEvt <- nil
}

// The id of the target node, i.e. the node whose state will be changed by the event executing.
func (fe FunctionEvent) Target() int {
	return fe.target
}
