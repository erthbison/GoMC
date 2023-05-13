package event

import (
	"fmt"
)

// Represent the target node crashing
type CrashEvent struct {
	target int
	crash  func(int) error

	id EventId
}

// Create a CrashEvent
//
// target is the id of the target node.
// crash is a function that will be called when the event is executed.
func NewCrashEvent(target int, crash func(int) error) CrashEvent {
	return CrashEvent{
		target: target,
		crash:  crash,

		id: EventId(fmt.Sprint("Crash", target)),
	}
}

// An id that identifies the event.
// Two events that provided the same input state results in the same output state should have the same id
//
// New event implementations should include a identifier of the event type to prevent accidental collisions with other implementations
func (ce CrashEvent) Id() EventId {
	return ce.id
}

// A method executing the event.
//
// Call the crash function with the target id
//
// The event will be executed on a separate goroutine.
// It should signal on the channel if it is clear for the simulator to proceed to processing of the state and the next event.
// Panics raised while executing the event is recovered by the simulator and returned as errors
func (ce CrashEvent) Execute(_ any, evtChan chan error) {
	evtChan <- ce.crash(ce.target)
}

// The id of the target node, i.e. the node whose state will be changed by the event executing.
func (ce CrashEvent) Target() int {
	return ce.target
}

func (ce CrashEvent) String() string {
	return fmt.Sprintf("{Crash Target: %v}", ce.target)
}
