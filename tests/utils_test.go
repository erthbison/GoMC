package gomc_test

import (
	"gomc"
	"gomc/event"
	"strconv"
)

// Create some dummy types and states for use when testing
type MockNode struct {
	Id int
}

func (n *MockNode) Foo(from, to int, msg []byte) {}
func (n *MockNode) Bar(from, to int, msg []byte) {}

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

func (ms *MockStateManager) NewRun() *gomc.RunStateManager[MockNode, State] {
	return &gomc.RunStateManager[MockNode, State]{}
}

func (ms *MockStateManager) State() gomc.StateSpace[State] {
	return gomc.TreeStateSpace[State]{}
}

func (ms *MockStateManager) AddRun([]gomc.GlobalState[State]) {
}

type MockEvent struct {
	id       int
	target   int
	executed bool
}

func (me MockEvent) Id() string {
	return strconv.Itoa(me.id)
}

func (me MockEvent) Execute(_ any, chn chan error) {
	chn <- nil
}

func (me MockEvent) Target() int {
	return me.target
}
