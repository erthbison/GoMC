package gomc_test

import (
	"gomc"
	"gomc/event"
	"gomc/scheduler"
	"strconv"
)

// Create some dummy types and states for use when testing
type MockNode struct {
	Id int
}

func (n *MockNode) Foo(from, to int, msg []byte) {}
func (n *MockNode) Bar(from, to int, msg []byte) {}

type State struct{}

type MockGlobalScheduler struct{}

func NewMockGlobalScheduler() *MockGlobalScheduler {
	return &MockGlobalScheduler{}
}

func (mgs *MockGlobalScheduler) GetRunScheduler() scheduler.RunScheduler {
	return NewMockScheduler()
}

type MockRunScheduler struct {
	inEvent  chan event.Event
	outEvent chan event.Event
	endRun   chan interface{}
}

func NewMockScheduler() *MockRunScheduler {
	return &MockRunScheduler{
		make(chan event.Event),
		make(chan event.Event),
		make(chan interface{}),
	}
}

func (ms *MockRunScheduler) AddEvent(evt event.Event) {
	ms.inEvent <- evt
}

func (ms *MockRunScheduler) GetEvent() (event.Event, error) {
	return <-ms.outEvent, nil
}

func (ms *MockRunScheduler) StartRun() error {
	return nil
}

func (ms *MockRunScheduler) EndRun() {
	ms.endRun <- nil
}

type MockStateManager struct {
}

func NewMockStateManager() *MockStateManager {
	return &MockStateManager{}
}

func (ms *MockStateManager) GetRunStateManager() *gomc.RunStateManager[MockNode, State] {
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
