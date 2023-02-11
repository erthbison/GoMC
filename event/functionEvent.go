package event

import "fmt"

// An function provided by the users.
// Is used to start the execution of the algorithm
type FunctionEvent[T any] struct {
	// Unique id that is used to identify the event.
	// Since the functions are provided in sequential order at the start of the run this will be consistent between runs
	index  int
	target int
	f      func(*T) error
}

func NewFunctionEvent[T any](i int, target int, f func(*T) error) FunctionEvent[T] {
	return FunctionEvent[T]{
		index:  i,
		target: target,
		f:      f,
	}
}

func (fe FunctionEvent[T]) Id() string {
	return fmt.Sprintf("Function %v", fe.index)
}

func (fe FunctionEvent[T]) String() string {
	return fmt.Sprintf("{Function %v}", fe.index)
}

func (fe FunctionEvent[T]) Execute(node *T, nextEvt chan error) {
	nextEvt <- fe.f(node)
}

func (fe FunctionEvent[T]) Target() int {
	return fe.target
}
