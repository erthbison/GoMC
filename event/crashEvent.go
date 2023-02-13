package event

import (
	"fmt"
)

type CrashEvent[T any] struct {
	target int
	crash  func(int)
}

func NewCrashEvent[T any](target int, crash func(int)) CrashEvent[T] {
	return CrashEvent[T]{
		target: target,
		crash:  crash,
	}
}

// An id that identifies the event. Two events that provided the same input state results in the same output state should have the same id
func (ce CrashEvent[T]) Id() string {
	return fmt.Sprintf("Crash Target %v", ce.target)
}

// A method executing the event. The event will be executed on a separate goroutine.
// It should signal on the channel if it is clear for the simulator to proceed to processing of the state and the next event.
// An event should be able to be executed multiple times and any two events with the same Id should be interchangeable.
// I.e. it does not matter which of the events you call the Execute method on. The results should be the same
// Panics raised while executing the event is recovered by the simulator and returned as errors
func (ce CrashEvent[T]) Execute(_ *T, evtChan chan error) {
	ce.crash(ce.target)
	evtChan <- nil
}

// The id of the target node, i.e. the node whose state will be changed by the event executing.
// Is used to identify if an event is still enabled, or if it has been disabled, e.g. because the node crashed.
func (ce CrashEvent[T]) Target() int {
	return ce.target
}

func (ce CrashEvent[T]) String() string {
	return fmt.Sprintf("{Crash Target: %v}", ce.target)
}
