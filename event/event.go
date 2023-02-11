package event

// An event represents some kind of action that will be scheduled and interleaved by the simulator
// They contain the information to execute themselves on some generic node and contain an identifier
type Event[T any] interface {
	// An id that identifies the event. Two events that provided the same input state results in the same output state should have the same id
	Id() string

	// A method executing the event. The event will be executed on a separate goroutine.
	// It should signal on the channel if it is clear for the simulator to proceed to processing of the state and the next event.
	// An event should be able to be executed multiple times and any two events with the same Id should be interchangeable.
	// I.e. it does not matter which of the events you call the Execute method on. The results should be the same
	// Panics raised while executing the event is recovered by the simulator and returned as errors
	Execute(map[int]*T, chan error)

	// The id of the target node, i.e. the node whose state will be changed by the event executing.
	// Is used to identify if an event is still enabled, or if it has been disabled, e.g. because the node crashed.
	Target() int
}

func EventsEquals[T any](a, b Event[T]) bool {
	return a.Id() == b.Id()
}
