package event

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

type node struct {
	foo bool
	bar bool
}

func (n *node) Foo(from int, msg []byte) {
	n.foo = true
}
func (n *node) Bar(from int, msg []byte) {
	n.bar = true
}

func TestFunctionEvent(t *testing.T) {
	// Call a function event of a valid node with valid input
	// Check tht we receive a message on the error channel and that the foo flag on the node is true
	evt := NewFunctionEvent(0, 0, "Foo",
		reflect.ValueOf(0),
		reflect.ValueOf([]byte("Foo")),
	)
	n := &node{}
	errChan := make(chan error)
	go func() {
		evt.Execute(n, errChan)
	}()
	select {
	case val := <-errChan:
		if val != nil {
			t.Errorf("Expected to receive no error when executing event. Got: %v", val)
		}
	case <-time.After(5 * time.Second):
		t.Errorf("Expected to receive a message on the errChan")
	}
	if !n.foo {
		t.Errorf("Expected the Foo function to have been called and the foo flag to be true")
	}
}

func TestMessageEvent(t *testing.T) {
	evt := NewMessageEvent(0, 1, "Bar", 0, []byte("Bar"))
	n := &node{}
	errChan := make(chan error)
	go func() {
		evt.Execute(n, errChan)
	}()
	select {
	case val := <-errChan:
		if val != nil {
			t.Errorf("Expected to receive no error when executing event. Got: %v", val)
		}
	case <-time.After(5 * time.Second):
		t.Errorf("Expected to receive a message on the errChan")
	}
	if !n.bar {
		t.Errorf("Expected the Bar function to have been called and the Bar flag to be true")
	}
}

func TestCrashEvent(t *testing.T) {
	crashedNode := 0
	evt := NewCrashEvent(5, func(i int) error { crashedNode = i; return nil })
	n := &node{}
	errChan := make(chan error)
	go func() {
		evt.Execute(n, errChan)
	}()
	select {
	case val := <-errChan:
		if val != nil {
			t.Errorf("Expected to receive no error when executing event. Got: %v", val)
		}
	case <-time.After(5 * time.Second):
		t.Errorf("Expected to receive a message on the errChan")
	}
	if crashedNode != 5 {
		t.Errorf("Expected the provided function to be executed")
	}
}

func TestSleepEvent(t *testing.T) {
	// Execute a sleep event. Test that a message is sent on the timeChan ensuring that the sleep ends.
	// Also check that we do not receive a message on the errorChan
	sleepChan := make(chan time.Time)
	foo := NewSleepEvent("Foo", 0, sleepChan)
	errChan := make(chan error)
	go foo.Execute(&node{}, errChan)
	select {
	case <-sleepChan:
	case <-time.After(5 * time.Second):
		t.Errorf("Expected the event to execute and end the sleep")
	}

	select {
	case <-errChan:
		t.Errorf("Did not expect to receive a message on the errorChan")
	default:
	}
}

func TestSleepEventOnSameLocation(t *testing.T) {
	// Test that two Sleep events with the same id are interchangeable
	sleepChan1 := make(chan time.Time)
	sleepChan2 := make(chan time.Time)
	foo1 := NewSleepEvent("Foo", 0, sleepChan1)
	foo2 := NewSleepEvent("Foo", 0, sleepChan2)
	if foo1.Id() != foo2.Id() {
		t.Errorf("Expected the Ids to be the same")
	}
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		<-sleepChan1
		wg.Done()
	}()
	go func() {
		<-sleepChan2
		wg.Done()
	}()
	errChan := make(chan error)
	foo1.Execute(&node{}, errChan)
	foo2.Execute(&node{}, errChan)
	wg.Wait()
}
