package runner

import "fmt"

type Record interface {
	Target() int
	fmt.Stringer
}

// Sent by a node after executing some event (received a message or executed local event)
type StateRecord[S any] struct {
	target int
	state  S
}

func (sr StateRecord[S]) Target() int {
	return sr.target
}

func (er StateRecord[S]) String() string {
	return fmt.Sprintf("[State - %+v]", er.state)
}

// Sent when a node execute some local event
type ExecutionRecord struct {
	target int
	evt    interface{}
}

func (er ExecutionRecord) Target() int {
	return er.target
}

func (er ExecutionRecord) String() string {
	return fmt.Sprintf("[Execution - Evt %v]", er.evt)
}

// Sent when a node either send or receive a message.
// The sent flag is true if the message was sent by the node and false if it was received
type MessageRecord struct {
	from, to int
	sent     bool
	evt      interface{}
}

func (m MessageRecord) Target() int {
	if m.sent {
		return m.from
	}
	return m.to
}

func (m MessageRecord) String() string {
	var t string
	if m.sent {
		t = "Sent"
	} else {
		t = "Received"
	}
	return fmt.Sprintf("[Message %v - %v]", t, m.evt)
}
