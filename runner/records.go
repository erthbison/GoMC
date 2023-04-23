package runner

import "fmt"

type Record interface {
	Target() int
	fmt.Stringer
}

type StateRecord struct {
	target int
	state  interface{}
}

func (sr StateRecord) Target() int {
	return sr.target
}

func (er StateRecord) String() string {
	return fmt.Sprintf("[State - %+v]", er.state)
}

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
