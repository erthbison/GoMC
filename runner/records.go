package runner

import (
	"fmt"
	"gomc/event"
)

type Record interface {
	Target() int
	fmt.Stringer
}

// Sent by a node after executing some event (received a message or executed local event)
type StateRecord[S any] struct {
	target int
	State  S
}

func (sr StateRecord[S]) Target() int {
	return sr.target
}

func (er StateRecord[S]) String() string {
	return fmt.Sprintf("[State - %+v]", er.State)
}

// Sent when a node execute some local event
type ExecutionRecord struct {
	target int
	Evt    event.Event
}

func (er ExecutionRecord) Target() int {
	return er.target
}

func (er ExecutionRecord) String() string {
	return fmt.Sprintf("[Execution - Evt %v]", er.Evt)
}

// Sent when a node either send or receive a message.
// The sent flag is true if the message was sent by the node and false if it was received
type MessageRecord struct {
	From, To int
	Sent     bool
	Evt      event.MessageEvent
}

func (m MessageRecord) Target() int {
	if m.Sent {
		return m.From
	}
	return m.To
}

func (m MessageRecord) String() string {
	var t string
	if m.Sent {
		t = "Sent"
	} else {
		t = "Received"
	}
	return fmt.Sprintf("[Message %v - %v]", t, m.Evt)
}
