package event

import (
	"fmt"
	"reflect"
)

// An event representing the arrival of a message on a node.
// Calls the function specified with the Type parameter when executed
type MessageEvent struct {
	From   int
	To     int
	Type   string
	Params []reflect.Value

	id EventId
}

func NewMessageEvent(from, to int, msgType string, params ...any) MessageEvent {
	valueParams := make([]reflect.Value, len(params))
	for i, val := range params {
		valueParams[i] = reflect.ValueOf(val)
	}
	return MessageEvent{
		From:   from,
		To:     to,
		Type:   msgType,
		Params: valueParams,

		id: EventId(fmt.Sprint("Message ", from, to, msgType, params)),
	}
}

func (me MessageEvent) Id() EventId {
	return me.id
}

func (me MessageEvent) String() string {
	return fmt.Sprintf("{From: %v, To: %v, Type: %s}", me.From, me.To, me.Type)
}

func (me MessageEvent) Execute(node any, nextEvt chan error) {
	// Use reflection to call the specified method on the node
	method := reflect.ValueOf(node).MethodByName(me.Type)
	method.Call(me.Params)
	nextEvt <- nil
}

func (me MessageEvent) Target() int {
	return me.To
}
