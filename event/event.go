package event

// An event represents some kind of action that will be scheduled and interleaved by the simulator
// They contain the information to execute themselves on some generic node and contain an identifier
type Event interface {
	// An id that identifies the event.
	// Two events that provided the same input state results in the same output state should have the same id
	//
	// New event implementations should include a identifier of the event type to prevent accidental collisions with other implementations
	Id() EventId

	// A method executing the event.
	// The event will be executed on a separate goroutine.
	// It should signal on the channel if it is clear for the simulator to proceed to processing of the state and the next event.
	// Panics raised while executing the event is recovered by the simulator and returned as errors
	Execute(node any, errorChan chan error)

	// The id of the target node, i.e. the node whose state will be changed by the event executing.
	Target() int
}

// An event that is used to represent some kind of message between two nodes.
type MessageEvent interface {
	Event

	// Returns the id of the node receiving the event
	To() int
	// Returns the id of the node sending the event
	From() int
}

// Compares two events
//
// Returns true of both events have the same id or if both are nil.
// Returns false otherwise.
func EventsEquals(a, b Event) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Id() == b.Id()
}

// An id that identifies the event.
// Two events that provided the same input state results in the same output state should have the same id
type EventId string
