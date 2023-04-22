package event

import (
	"fmt"
	"reflect"
)

// An event representing the arrival of a message on a node.
// Calls the function specified with the Type parameter when executed
type MessageHandlerEvent struct {
	from    int
	to      int
	msgType string
	params  []reflect.Value

	id EventId
}

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

func (me MessageHandlerEvent) Id() EventId {
	return me.id
}

func (me MessageHandlerEvent) String() string {
	return fmt.Sprintf("{From: %v, To: %v, Type: %s}", me.from, me.to, me.msgType)
}

func (me MessageHandlerEvent) Execute(node any, nextEvt chan error) {
	// Use reflection to call the specified method on the node
	method := reflect.ValueOf(node).MethodByName(me.msgType)
	method.Call(me.params)
	nextEvt <- nil
}

func (me MessageHandlerEvent) Target() int {
	return me.to
}

func (me MessageHandlerEvent) To() int {
	return me.to
}

func (me MessageHandlerEvent) From() int {
	return me.from
}
