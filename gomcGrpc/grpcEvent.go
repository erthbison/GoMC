package gomcGrpc

import "fmt"

type GrpcEvent struct {
	from   int
	target int
	method string
	msg    any
	wait   chan bool
}

// An id that identifies the event. Two events that provided the same input state results in the same output state should have the same id
func (ge GrpcEvent) Id() string {
	return fmt.Sprintf("GrpcEvent From: %v To: %v Method: %v Msg %v", ge.from, ge.target, ge.method, ge.msg)
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
