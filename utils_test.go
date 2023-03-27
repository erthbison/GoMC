package gomc

import (
	"gomc/event"
	"gomc/scheduler"
	"strconv"
)

// Create some dummy types and states for use when testing
type MockNode struct {
	Id int

	crashed bool
	val     int
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
	return NewMockRunScheduler()
}

type MockRunScheduler struct {
	eventQueue  []event.Event
	addedEvents []event.Event
	index       int
	runEnded    bool
}

func NewMockRunScheduler(events ...event.Event) *MockRunScheduler {
	return &MockRunScheduler{
		eventQueue:  events,
		index:       0,
		addedEvents: make([]event.Event, 0),
	}
}

func (ms *MockRunScheduler) AddEvent(evt event.Event) {
	ms.addedEvents = append(ms.addedEvents, evt)
}

func (ms *MockRunScheduler) GetEvent() (event.Event, error) {
	if ms.index < len(ms.eventQueue) {
		evt := ms.eventQueue[ms.index]
		ms.index++
		return evt, nil
	}
	return nil, scheduler.RunEndedError
}

func (ms *MockRunScheduler) StartRun() error {
	if ms.runEnded {
		return scheduler.RunEndedError
	}
	return nil
}

func (ms *MockRunScheduler) EndRun() {
	ms.addedEvents = make([]event.Event, 0)
}

type MockStateManager struct {
	receivedRun []GlobalState[State]
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

func (ms *MockStateManager) AddRun(run []GlobalState[State]) {
	ms.receivedRun = run
}

type MockEvent struct {
	id       int
	target   int
	executed bool
	val      int
}

func (me MockEvent) Id() string {
	return strconv.Itoa(me.id)
}

func (me MockEvent) Execute(n any, chn chan error) {
	if me.val == -1 {
		panic("Node panicked during testing")
	}
	tmp := n.(*MockNode)
	if !tmp.crashed {
		tmp.val = me.val
	}
	chn <- nil
}

func (me MockEvent) Target() int {
	return me.target
}
