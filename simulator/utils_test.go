package simulator

import (
	"gomc/event"
	"gomc/eventManager"
	"gomc/failureManager"
	"gomc/scheduler"
	"gomc/state"
	"gomc/stateManager"
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

func (mgs *MockGlobalScheduler) Reset() {
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
	receivedRun []state.GlobalState[State]
}

func NewMockStateManager() *MockStateManager {
	return &MockStateManager{}
}

func (ms *MockStateManager) GetRunStateManager() *stateManager.RunStateManager[MockNode, State] {
	return &stateManager.RunStateManager[MockNode, State]{}
}

func (ms *MockStateManager) State() state.StateSpace[State] {
	return state.TreeStateSpace[State]{}
}

func (ms *MockStateManager) AddRun(run []state.GlobalState[State]) {
	ms.receivedRun = run
}

func (ms *MockStateManager) Reset() {
	ms.receivedRun = nil
}

type MockEvent struct {
	id       event.EventId
	target   int
	executed bool
	val      int
}

func (me MockEvent) Id() event.EventId {
	return me.id
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

type MockFailureManager struct {
	failingNodes []int
	crashFunc    func(*MockNode)
}

func (mfm *MockFailureManager) GetRunFailureManager(ea eventManager.EventAdder) failureManager.RunFailureManager[MockNode] {
	return NewMockRunFailureManager(
		ea, mfm.failingNodes, mfm.crashFunc,
	)
}

func NewMockFailureManager(failingNodes []int, crashFunc func(*MockNode)) *MockFailureManager {
	return &MockFailureManager{
		failingNodes: failingNodes,
		crashFunc:    crashFunc,
	}
}

type MockRunFailureManager[T any] struct {
	ea           eventManager.EventAdder
	failingNodes []int
	crashFunc    func(*T)

	correct map[int]bool
}

func NewMockRunFailureManager(ea eventManager.EventAdder, failingNodes []int, crashFunc func(*MockNode)) *MockRunFailureManager[MockNode] {
	return &MockRunFailureManager[MockNode]{
		ea:           ea,
		failingNodes: failingNodes,
		crashFunc:    crashFunc,
	}
}

func (mrfm *MockRunFailureManager[MockNode]) Init(nodes map[int]*MockNode) {
	for _, id := range mrfm.failingNodes {
		mrfm.ea.AddEvent(event.NewCrashEvent(id, mrfm.nodeCrash))
	}
}

func (mrfm *MockRunFailureManager[MockNode]) CorrectNodes() map[int]bool {
	return mrfm.correct
}

func (mrfm *MockRunFailureManager[MockNode]) Subscribe(nodeId int, callback func(id int, status bool)) {

}

func (mrfm *MockRunFailureManager[T]) nodeCrash(id int) error {
	mrfm.correct[id] = false
	return nil
}
