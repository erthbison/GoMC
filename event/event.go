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
}

func EventsEquals[T any](a, b Event[T]) bool {
	return a.Id() == b.Id()
}

// An event representing the start of the simulation.
// The execution of a start event performs no change to the state
type StartEvent[T any] struct{}

func (se StartEvent[T]) Id() string {
	return "Start"
}

func (se StartEvent[T]) String() string {
	return "{Start}"
}

func (se StartEvent[T]) Execute(_ map[int]*T, _ chan error) {}

// An event representing the end of a run
// The execution of a start event performs no change to the state
type EndEvent[T any] struct{}

func (ee EndEvent[T]) Id() string {
	return "End"
}

func (ee EndEvent[T]) String() string {
	return "{End}"
}

func (ee EndEvent[T]) Execute(_ map[int]*T, _ chan error) {}
