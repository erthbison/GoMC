package event

import (
	"fmt"
	"time"
)

// An event representing a timeout.
//
// It is analogous to time.Sleep and can be used to represent timeout
type SleepEvent struct {
	caller      string
	target      int // The id of the target node
	timeoutChan chan time.Time

	id EventId
}

// Create a SleepEvent
//
// caller is a string representing the location in the code where the timeout was called. It is used to create the id of the node.
// target is the id of the node that called the timeout.
// timeoutChan is the channel that is waiting for the timeout to expire
func NewSleepEvent(caller string, target int, timeoutChan chan time.Time) SleepEvent {
	evt := SleepEvent{
		caller:      caller,
		target:      target,
		timeoutChan: timeoutChan,

		id: EventId(fmt.Sprint("Sleep ", target, caller)),
	}
	return evt
}

// An id that identifies the event.
// Two events that provided the same input state results in the same output state should have the same id
//
// New event implementations should include a identifier of the event type to prevent accidental collisions with other implementations
func (se SleepEvent) Id() EventId {
	return se.id
}

func (se SleepEvent) String() string {
	return fmt.Sprintf("{Sleep Target: %v}", se.target)
}

// A method executing the event.
// The event will be executed on a separate goroutine.
// It should signal on the channel if it is clear for the simulator to proceed to processing of the state and the next event.
// Panics raised while executing the event is recovered by the simulator and returned as errors
func (se SleepEvent) Execute(node any, _ chan error) {
	// Send a signal on the timeout channel
	// Don't signal on the error channel since the event that was paused by the sleep event will continue running and the simulator can therefore not begin collecting state yet.
	// The event that continues after the sleep event will signal to the simulator when it has completed.
	se.timeoutChan <- time.Time{}
}

// The id of the target node, i.e. the node whose state will be changed by the event executing.
func (se SleepEvent) Target() int {
	return se.target
}
