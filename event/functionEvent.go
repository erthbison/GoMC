package event

import "fmt"

// An function provided by the users.
// Is used to start the execution of the algorithm
type FunctionEvent[T any] struct {
	// Unique id that is used to identify the event.
	// Since the functions are provided in sequential order at the start of the run this will be consistent between runs
	Index int
	F     func(map[int]*T) error
}

func (fe FunctionEvent[T]) Id() string {
	return fmt.Sprintf("Function %v", fe.Index)
}

func (fe FunctionEvent[T]) String() string {
	return fmt.Sprintf("{Function %v}", fe.Index)
}

func (fe FunctionEvent[T]) Execute(node map[int]*T, nextEvt chan error) {
	nextEvt <- fe.F(node)
}
