package runner

import (
	"fmt"
	"gomc/event"
)

// A record that has been collected by the Runner.
//
// The Target method returns the id of the node where the event originates.
// There are three types of records:
//
// StateRecord: Which contains the state of a node after an event has been executed.
//
// ExecutionRecord: Which represent the execution of a local event.
//
// MessageRecord: Which are created when a message is sent and when it is executed on a node.
type Record interface {
	Target() int
	fmt.Stringer
}

// Sent by a node after executing some event (received a message or executed local event)
type StateRecord[S any] struct {
	target int
	State  S
}

// The id of the node whose state was collected
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

// The id of the node where the local event was executed
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
	// Sent is true if the record was created because a node sent this MessageEvent
	// It is false if the record was created because the MessageEvent was executed on a node.
	Sent bool
	Evt  event.MessageEvent
}

// The id of the node
//
// If the message was sent by a node, it is the id of the event that sent the message.
// if the message was received on a node it is the id of the event that received the message.
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
