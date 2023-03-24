package gomc

import (
	"gomc/event"
	"gomc/scheduler"
	"strconv"
)

// Create some dummy types and states for use when testing
type MockNode struct {
	Id int

	val int
}

func (n *MockNode) UpdateVal(val int) {
	n.val = val
}

func GetState(n *MockNode) State {
	return State{
		val: n.val,
	}
}

type State struct {
	val int
}

type MockGlobalScheduler struct{}

func NewMockGlobalScheduler() *MockGlobalScheduler {
	return &MockGlobalScheduler{}
}

func (mgs *MockGlobalScheduler) GetRunScheduler() scheduler.RunScheduler {
	return NewMockScheduler()
}

type MockRunScheduler struct {
	eventQueue []event.Event
	index      int
}

func NewMockScheduler(events ...event.Event) *MockRunScheduler {
	return &MockRunScheduler{
		eventQueue: events,
		index:      0,
	}
}

func (ms *MockRunScheduler) AddEvent(evt event.Event) {

}

func (ms *MockRunScheduler) GetEvent() (event.Event, error) {
	if ms.index < len(ms.eventQueue) {
		return ms.eventQueue[ms.index], nil
	}
	return nil, scheduler.RunEndedError
}

func (ms *MockRunScheduler) StartRun() error {
	return nil
}

func (ms *MockRunScheduler) EndRun() {
	
}

type MockStateManager struct {
}

func NewMockStateManager() *MockStateManager {
	return &MockStateManager{}
}

func (ms *MockStateManager) GetRunStateManager() *RunStateManager[MockNode, State] {
	return &RunStateManager[MockNode, State]{}
}

func (ms *MockStateManager) State() StateSpace[State] {
	return treeStateSpace[State]{}
}

func (ms *MockStateManager) AddRun([]GlobalState[State]) {
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
