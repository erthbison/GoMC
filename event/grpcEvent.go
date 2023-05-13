package event

import (
	"fmt"
)

// An Event representing a node sending a message using asynchronous gRPC request.
//
// An asynchronous RPC call is one that is made in a separate goroutine, and where there response is ignored.
// e.g. go ExampleRpcServer.Foo(...)
//
// The message that is withheld by an interceptor.
// When the event is executed the message is released.
type GrpcEvent struct {
	from   int
	target int
	method string
	wait   chan bool

	id EventId
}

// Create a new GrpcEvent
//
// from is teh id of the node sending the message, to is the id of the node receiving it.
// method is a string representation of the msg used.
// msg is the message sent.
// wait is a channel that will be used to represent that the message can be sent to the target node.
func NewGrpcEvent(from int, to int, method string, msg interface{}, wait chan bool) GrpcEvent {
	return GrpcEvent{
		target: to,
		from:   from,
		method: method,
		wait:   wait,

		id: EventId(fmt.Sprint("GrpcEvent", from, to, method, msg)),
	}
}

// An id that identifies the event.
// Two events that provided the same input state results in the same output state should have the same id
//
// New event implementations should include a identifier of the event type to prevent accidental collisions with other implementations
func (ge GrpcEvent) Id() EventId {
	return ge.id
}

// A method executing the event.
//
// Release the message so that it is sent to the target node
//
// The event will be executed on a separate goroutine.
// It should signal on the channel if it is clear for the simulator to proceed to processing of the state and the next event.
// Panics raised while executing the event is recovered by the simulator and returned as errors
func (ge GrpcEvent) Execute(node any, errorChan chan error) {
	ge.wait <- true
}

// The id of the target node, i.e. the node whose state will be changed by the event executing.
func (ge GrpcEvent) Target() int {
	return ge.target
}

func (ge GrpcEvent) String() string {
	return fmt.Sprintf("GrpcEvent From: %v To: %v Method: %v", ge.from, ge.target, ge.method)
}

// Returns the id of the node receiving the event
func (ge GrpcEvent) To() int {
	return ge.target
}

// Returns the id of the node sending the event
func (ge GrpcEvent) From() int {
	return ge.from
}
