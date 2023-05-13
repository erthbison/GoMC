package event

import (
	"fmt"
	"reflect"
)

// An event representing the arrival of a message on a node.
//
// Does not incorporate any message passing mechanisms, instead it calls an message handler on the target node.
// Assumes that there exist some message handler of the node which can be called.
// Implements the MessageEvent interface
type MessageHandlerEvent struct {
	from    int
	to      int
	msgType string
	params  []reflect.Value

	id EventId
}

// Creates a MessageHandlerEvent
//
// from is the id of the sending node, to is the id of the receiving node.
// msgType is the name of the message handler method that will be called,
// params is the parameters that will be passed to the method.
func NewMessageHandlerEvent(from, to int, msgType string, params ...any) MessageHandlerEvent {
	valueParams := make([]reflect.Value, len(params))
	for i, val := range params {
		valueParams[i] = reflect.ValueOf(val)
	}
	return MessageHandlerEvent{
		from:    from,
		to:      to,
		msgType: msgType,
		params:  valueParams,

		id: EventId(fmt.Sprint("Message ", from, to, msgType, params)),
	}
}

// An id that identifies the event.
// Two events that provided the same input state results in the same output state should have the same id
//
// New event implementations should include a identifier of the event type to prevent accidental collisions with other implementations
func (me MessageHandlerEvent) Id() EventId {
	return me.id
}

func (me MessageHandlerEvent) String() string {
	return fmt.Sprintf("{From: %v, To: %v, Type: %s}", me.from, me.to, me.msgType)
}

// A method executing the event.
// The event will be executed on a separate goroutine.
// It should signal on the channel if it is clear for the simulator to proceed to processing of the state and the next event.
// Panics raised while executing the event is recovered by the simulator and returned as errors
//
// Calls the specified method on the target node with the provided parameters using reflection.
func (me MessageHandlerEvent) Execute(node any, nextEvt chan error) {
	// Use reflection to call the specified method on the node
	method := reflect.ValueOf(node).MethodByName(me.msgType)
	method.Call(me.params)
	nextEvt <- nil
}

// The id of the target node, i.e. the node whose state will be changed by the event executing.
func (me MessageHandlerEvent) Target() int {
	return me.to
}

// The id of the Node that the message is sent to
func (me MessageHandlerEvent) To() int {
	return me.to
}

// The id of the Node that the message is sent from
func (me MessageHandlerEvent) From() int {
	return me.from
}
