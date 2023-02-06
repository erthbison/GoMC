package gomc

import (
	"fmt"
	"reflect"
	"time"
)

type Event[T any] interface {
	// An id that identifies the event. Two events that provided the same input state results in the same output state should have the same id
	Id() string
	Execute(map[int]*T, chan error)
}

func EventsEquals[T any](a, b Event[T]) bool {
	return a.Id() == b.Id()
}

type StartEvent[T any] struct{}

func (se StartEvent[T]) Id() string {
	return "Start"
}

func (se StartEvent[T]) String() string {
	return "{Start}"
}

func (se StartEvent[T]) Execute(_ map[int]*T, _ chan error) {}

type EndEvent[T any] struct{}

func (ee EndEvent[T]) Id() string {
	return "End"
}

func (ee EndEvent[T]) String() string {
	return "{End}"
}

func (ee EndEvent[T]) Execute(_ map[int]*T, _ chan error) {}

type MessageEvent[T any] struct {
	From  int
	To    int
	Type  string
	Value []byte
}

func (me MessageEvent[T]) Id() string {
	return fmt.Sprintf("Message From: %v, To: %v, Type: %v, Value: %v", me.From, me.To, me.Type, me.Value)
}

func (me MessageEvent[T]) String() string {
	return fmt.Sprintf("{From: %v, To: %v, Type: %s}", me.From, me.To, me.Type)
}

func (me MessageEvent[T]) Execute(nodes map[int]*T, nextEvt chan error) {
	// Use reflection to call the specified method on the node
	node := nodes[me.To]
	method := reflect.ValueOf(node).MethodByName(me.Type)
	method.Call([]reflect.Value{
		reflect.ValueOf(me.From),
		reflect.ValueOf(me.To),
		reflect.ValueOf(me.Value),
	})
	nextEvt <- nil
}

type FunctionEvent[T any] struct {
	// Unique id that is used to identify the event.
	// Since the functions are provided in sequential order at the start of the run this will be consistent between runs
	index int
	F     func(map[int]*T) error
}

func (fe FunctionEvent[T]) Id() string {
	return fmt.Sprintf("Function %v", fe.index)
}

func (fe FunctionEvent[T]) String() string {
	return fmt.Sprintf("{Function %v}", fe.index)
}

func (fe FunctionEvent[T]) Execute(node map[int]*T, nextEvt chan error) {
	nextEvt <- fe.F(node)
}

type TimeoutEvent[T any] struct {
	caller      string
	timeoutChan map[string]chan time.Time
}

func (te TimeoutEvent[T]) Id() string {
	return fmt.Sprintf("Timeout Caller: %v", te.caller)
}

func (te TimeoutEvent[T]) String() string {
	return fmt.Sprintf("{Timeout Caller: %v}", te.caller)
}

func (te TimeoutEvent[T]) Execute(node map[int]*T, _ chan error) {
	timeout := te.timeoutChan[te.Id()]
	timeout <- time.Time{}
	close(timeout)
}

func NewTimeoutEvent[T any](caller string, timeoutChan map[string]chan time.Time) TimeoutEvent[T] {
	waitChan := make(chan time.Time)
	evt := TimeoutEvent[T]{
		caller:      caller,
		timeoutChan: timeoutChan,
	}
	timeoutChan[evt.Id()] = waitChan
	return evt
}
