package event

import "hash/maphash"

// An event represents some kind of action that will be scheduled and interleaved by the simulator
// They contain the information to execute themselves on some generic node and contain an identifier
type Event interface {
	// An id that identifies the event. Two events that provided the same input state results in the same output state should have the same id
	Id() uint64

	// A method executing the event. The event will be executed on a separate goroutine.
	// It should signal on the channel if it is clear for the simulator to proceed to processing of the state and the next event.
	// Panics raised while executing the event is recovered by the simulator and returned as errors
	Execute(node any, errorChan chan error)

	// The id of the target node, i.e. the node whose state will be changed by the event executing.
	// Is used to identify if an event is still enabled, or if it has been disabled, e.g. because the node crashed.
	Target() int
}

func EventsEquals(a, b Event) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Id() == b.Id()
}

var EventHash = maphash.MakeSeed()
