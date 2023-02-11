package gomc_test

import (
	"gomc/event"
	"strconv"
)

// Create some dummy types and states for use when testing
type node struct{}

func (n *node) Foo(from, to int, msg []byte) {}
func (n *node) Bar(from, to int, msg []byte) {}

type state struct{}

type MockScheduler struct {
	inEvent  chan event.Event[node]
	outEvent chan event.Event[node]
	endRun   chan interface{}
}

func NewMockScheduler() *MockScheduler {
	return &MockScheduler{
		make(chan event.Event[node]),
		make(chan event.Event[node]),
		make(chan interface{}),
	}
}

func (ms *MockScheduler) AddEvent(evt event.Event[node]) {
	ms.inEvent <- evt
}

func (ms *MockScheduler) GetEvent() (event.Event[node], error) {
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

func (msm *MockStateManager) UpdateGlobalState(map[int]*node) {}

func (msm *MockStateManager) EndRun() {}

type MockEvent struct {
	id       int
	executed bool
}

func (me MockEvent) Id() string {
	return strconv.Itoa(me.id)
}

func (me MockEvent) Execute(_ map[int]*node, chn chan error) {
	me.executed = true
	chn <- nil
}
