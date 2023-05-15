package eventManager

import (
	"gomc/event"
)

// An Event Manager used to send messages between nodes
//
// Represent a message as a call to a message handler method on the target node.
// When the message arrived the method is called with the parameters.
type Sender struct {
	ea EventAdder
}

// Create a new Sender with the provided EventAdder
func NewSender(sp SimulationParameters) *Sender {
	return &Sender{ea: sp.EventAdder}
}

// Creates a send function that creates an event representing the message to be sent to the target node
// The SendFunc is called with the id of the node that will send the messages.
// Should create a new send function for each node.
//
// The returned send function is used to represent mechanism that sends a message to a node
// id represent the target of the message
// msgType is the message handler that will be called when the message arrive.
// params is the parameters that will be passed to the message handler method.
func (s *Sender) SendFunc(id int) func(int, string, ...any) {
	return func(to int, msgType string, params ...any) {
		s.ea.AddEvent(event.NewMessageHandlerEvent(id, to, msgType, params...))
	}
}
