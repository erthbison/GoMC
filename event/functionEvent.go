package event

import (
	"fmt"
	"hash/maphash"
	"reflect"
)

// An function provided by the users.
// Is used to start the execution of the algorithm
type FunctionEvent struct {
	// Unique id that is used to identify the event.
	// Since the functions are provided in sequential order at the start of the run this will be consistent between runs
	index  int
	target int
	method string
	params []reflect.Value

	id uint64
}

func NewFunctionEvent(i int, target int, method string, params ...reflect.Value) FunctionEvent {
	return FunctionEvent{
		index:  i,
		target: target,
		method: method,
		params: params,

		id: maphash.String(EventHashSeed, fmt.Sprint("Function", i)),
	}
}

func (fe FunctionEvent) Id() uint64 {
	return fe.id
}

func (fe FunctionEvent) String() string {
	return fmt.Sprintf("{Function %v. Target: %v}", fe.index, fe.target)
}

func (fe FunctionEvent) Execute(node any, nextEvt chan error) {
	method := reflect.ValueOf(node).MethodByName(fe.method)
	method.Call(fe.params)
	nextEvt <- nil
}

func (fe FunctionEvent) Target() int {
	return fe.target
}
