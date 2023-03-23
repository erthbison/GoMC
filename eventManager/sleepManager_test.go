package eventManager

import (
	"gomc/event"
	"testing"
	"time"
)

type MockScheduler struct {
	eventStack []event.Event
}

func NewMockScheduler() *MockScheduler {
	return &MockScheduler{
		eventStack: make([]event.Event, 0),
	}
}

func (ms *MockScheduler) AddEvent(evt event.Event) {
	ms.eventStack = append(ms.eventStack, evt)
}

func (ms *MockScheduler) GetEvent() (event.Event, error) {
	evt := ms.eventStack[len(ms.eventStack)-1]
	ms.eventStack = ms.eventStack[:len(ms.eventStack)-1]
	return evt, nil
}

func (ms *MockScheduler) StartRun() error { return nil }

func (ms *MockScheduler) EndRun() {}

func (ms *MockScheduler) NodeCrash(i int) {}

type Node struct{}

func (n *Node) Foo(from, to int, msg []byte) {}
func (n *Node) Bar(from, to int, msg []byte) {}

func TestSleepManager(t *testing.T) {
	sch := NewMockScheduler()
	nextEvent := make(chan error)
	sm := NewSleepManager(sch, nextEvent)
	sleep := sm.SleepFunc(0)
	notBlockedChan := make(chan bool)
	go func() {
		// Need to check that sleep actually blocks until the event is executed
		sleep(time.Second)
		notBlockedChan <- true
	}()

	// Wait until the nextEvent signal is received than execute the event.
	// Like the simulator would
	select {
	case <-nextEvent:
		evt, _ := sch.GetEvent()
		evt.Execute(&Node{}, nextEvent)
	case <-notBlockedChan:
		t.Fatalf("Sleep returned before the event was executed")
	}
	select {
	case <-notBlockedChan:
	case <-time.After(time.Second * 5):
		t.Fatalf("Sleep is still blocking after executing the event")
	}
}
