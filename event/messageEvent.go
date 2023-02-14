package event

import (
	"fmt"
	"reflect"
)

// An event representing the arrival of a message on a node.
// Calls the function specified with the Type parameter when executed
type MessageEvent struct {
	From  int
	To    int
	Type  string
	Value []byte

	id string
}

func NewMessageEvent(from, to int, msgType string, val []byte) MessageEvent {
	return MessageEvent{
		From:  from,
		To:    to,
		Type:  msgType,
		Value: val,
		id:    fmt.Sprintf("Message From: %v, To: %v, Type: %v, Value: %v", from, to, msgType, val),
	}
}

func (me MessageEvent) Id() string {
	return me.id
}

func (me MessageEvent) String() string {
	return fmt.Sprintf("{From: %v, To: %v, Type: %s}", me.From, me.To, me.Type)
}

func (me MessageEvent) Execute(node any, nextEvt chan error) {
	// Use reflection to call the specified method on the node
	method := reflect.ValueOf(node).MethodByName(me.Type)
	method.Call([]reflect.Value{
		reflect.ValueOf(me.From),
		reflect.ValueOf(me.Value),
	})
	nextEvt <- nil
}

func (me MessageEvent) Target() int {
	return me.To
}
