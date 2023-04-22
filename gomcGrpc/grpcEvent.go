package gomcGrpc

import (
	"fmt"
	"gomc/event"
)

type GrpcEvent struct {
	from   int
	target int
	method string
	wait   chan bool

	id event.EventId
}

func NewGrpcEvent(from int, to int, method string, msg interface{}, wait chan bool) GrpcEvent {
	return GrpcEvent{
		target: to,
		from:   from,
		method: method,
		wait:   wait,

		id: event.EventId(fmt.Sprint("GrpcEvent", from, to, method, msg)),
	}
}

// An id that identifies the event. Two events that provided the same input state results in the same output state should have the same id
func (ge GrpcEvent) Id() event.EventId {
	return ge.id
}

// A method executing the event. The event will be executed on a separate goroutine.
// It should signal on the channel if it is clear for the simulator to proceed to processing of the state and the next event.
// Panics raised while executing the event is recovered by the simulator and returned as errors
func (ge GrpcEvent) Execute(node any, errorChan chan error) {
	ge.wait <- true
}

// The id of the target node, i.e. the node whose state will be changed by the event executing.
// Is used to identify if an event is still enabled, or if it has been disabled, e.g. because the node crashed.
func (ge GrpcEvent) Target() int {
	return ge.target
}

func (ge GrpcEvent) String() string {
	return fmt.Sprintf("GrpcEvent From: %v To: %v Method: %v", ge.from, ge.target, ge.method)
}

func (ge GrpcEvent) To() int {
	return ge.target
}

func (ge GrpcEvent) From() int {
	return ge.from
}
