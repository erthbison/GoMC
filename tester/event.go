package tester

import (
	"fmt"
	"time"
)

type Event interface {
	// An id that identifies the event. Two events that provided the same input state results in the same output state should have the same id
	Id() string
}

func EventsEquals(a, b Event) bool {
	return a.Id() == b.Id()
}

type StartEvent struct{}

func (se StartEvent) Id() string {
	return "Start"
}

func (se StartEvent) String() string {
	return "{Start}"
}

type EndEvent struct{}

func (ee EndEvent) Id() string {
	return "End"
}

func (ee EndEvent) String() string {
	return "{End}"
}

type MessageEvent struct {
	From  int
	To    int
	Type  string
	Value []byte
}

func (me MessageEvent) Id() string {
	return fmt.Sprintf("Message From: %v, To: %v, Type: %v, Value: %v", me.From, me.To, me.Type, me.Value)
}

func (me MessageEvent) String() string {
	return fmt.Sprintf("{From: %v, To: %v, Type: %s}", me.From, me.To, me.Type)
}

type FunctionEvent[T any] struct {
	// Unique id that is used to identify the event.
	// Since the functions are provided in sequential order at the start of the run this will be consistent between runs
	index int
	F     func(map[int]*T)
}

func (fe FunctionEvent[T]) Id() string {
	return fmt.Sprintf("Function %v", fe.index)
}

func (fe FunctionEvent[T]) String() string {
	return fmt.Sprintf("{Function %v}", fe.index)
}

type TimeoutEvent struct {
	caller   string
	duration time.Time
	Timeout  chan<- time.Time
}
