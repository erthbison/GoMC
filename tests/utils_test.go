package gomc_test

import (
	"gomc/event"
	"strconv"
)

// Create some dummy types and states for use when testing
type Node struct{}

func (n *Node) Foo(from, to int, msg []byte) {}
func (n *Node) Bar(from, to int, msg []byte) {}

type State struct{}

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

type MockStateManager struct {
}

func NewMockStateManager() *MockStateManager {
	return &MockStateManager{}
}

func (msm *MockStateManager) UpdateGlobalState(map[int]*Node, map[int]bool, event.Event) {}

func (msm *MockStateManager) EndRun() {}

type MockEvent struct {
	id       int
	executed bool
}

func (me MockEvent) Id() string {
	return strconv.Itoa(me.id)
}

func (me MockEvent) Execute(_ map[int]*Node, chn chan error) {
	me.executed = true
	chn <- nil
}
