package gomc_test

import (
	"experimentation/gomc"
	"strconv"
)

// Create some dummy types and states for use when testing
type node struct{}

func (n *node) Foo(from, to int, msg []byte) {}
func (n *node) Bar(from, to int, msg []byte) {}

type state struct{}

type MockScheduler struct {
	inEvent  chan gomc.Event[node]
	outEvent chan gomc.Event[node]
	endRun   chan interface{}
}

func NewMockScheduler() *MockScheduler {
	return &MockScheduler{
		make(chan gomc.Event[node]),
		make(chan gomc.Event[node]),
		make(chan interface{}),
	}
}

func (ms *MockScheduler) AddEvent(evt gomc.Event[node]) {
	ms.inEvent <- evt
}

func (ms *MockScheduler) GetEvent() (gomc.Event[node], error) {
	return <-ms.outEvent, nil
}

func (ms *MockScheduler) EndRun() {
	ms.endRun <- nil
}

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
