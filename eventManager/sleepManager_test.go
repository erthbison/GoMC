package eventManager

import (
	"gomc/event"
	"testing"
	"time"
)

type MockScheduler struct {
	inEvent  chan event.Event
	outEvent chan event.Event
	endRun   chan interface{}
}

func NewMockScheduler() *MockScheduler {
	return &MockScheduler{
		make(chan event.Event),
		make(chan event.Event),
		make(chan interface{}),
	}
}

func (ms *MockScheduler) AddEvent(evt event.Event) {
	ms.inEvent <- evt
}

func (ms *MockScheduler) GetEvent() (event.Event, error) {
	return <-ms.outEvent, nil
}

func (ms *MockScheduler) EndRun() {
	ms.endRun <- nil
}

func (ms *MockScheduler) NodeCrash(i int) {}

type Node struct{}

func (n *Node) Foo(from, to int, msg []byte) {}
func (n *Node) Bar(from, to int, msg []byte) {}

func TestSleepManager(t *testing.T) {
	sch := NewMockScheduler()
	nextEvent := make(chan error)
	sm := NewSleepManager(sch, nextEvent)
	sleepFunc := sm.SleepFunc(0)
	notBlockedChan := make(chan bool)
	go func() {
		// Need to check that sleep actually blocks until the event is executed
		sleepFunc(time.Second)
		notBlockedChan <- true
	}()

	evt := <-sch.inEvent

	// Wait until the nextEvent signal is received than execute the event.
	// Like the simulator would
	select {
	case <-nextEvent:
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
