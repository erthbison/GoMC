package gomc

import (
	"fmt"
	"gomc/event"
	"gomc/stateManager"
	"reflect"
	"testing"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func TestSimulatorNoEvents(t *testing.T) {
	sch := NewMockGlobalScheduler()
	sm := NewMockStateManager()
	simulator := NewSimulator[MockNode, State](sch, false, false, 10000, 1000, 1)
	err := simulator.Simulate(
		sm,
		func(sp SimulationParameters) map[int]*MockNode {
			return map[int]*MockNode{0: {}}
		},
		[]int{},
		func(*MockNode) {},
	)
	if err == nil {
		t.Errorf("Expected to receive an error when not providing any functions to simulate")
	}

	sm = NewMockStateManager()
	err = simulator.Simulate(
		sm,
		func(sp SimulationParameters) map[int]*MockNode {
			return map[int]*MockNode{0: {}}
		},
		[]int{},
		func(*MockNode) {},
	)
	if err == nil {
		t.Errorf("Expected to receive an error when not providing any functions to simulate")
	}
}

func TestAddRequests(t *testing.T) {
	sch := NewMockRunScheduler()
	gsm := NewMockStateManager()
	sim := newRunSimulator(
		sch,
		stateManager.NewRunStateManager[MockNode, State](gsm, GetState),
		1000,
		false,
	)
	for i, test := range addRequestTests {
		err := sim.scheduleRequests(test.requests, test.nodes)

		isErr := (err != nil)
		if isErr != test.err {
			if isErr {
				t.Errorf("Test %v: Expected no error got: %v", i, err)
			} else {
				t.Errorf("Test %v: Expected to receive an error", i)
			}
			continue
		}

		if len(sch.addedEvents) != len(test.events) {
			t.Errorf("Test %v: Unexpected number of added events. Got %v. Expected %v.", i, len(sch.addedEvents), len(test.events))
		}

		for i, evt := range sch.addedEvents {

			expectedEvent := test.events[i]
			if _, ok := evt.(event.FunctionEvent); !ok {
				t.Errorf("Test %v: Added events of unexpected type. Expected: event.FunctionEvent. Got %T", i, evt)
			}
			if evt.Target() != expectedEvent.Target() {
				t.Errorf("Test %v: Unexpected target of added event. Got: %v. Expected: %v", i, evt.Target(), expectedEvent.Target())
			}
			if evt.Id() != expectedEvent.Id() {
				t.Errorf("Test %v: Unexpected id of added event. Got %v. Expected %v", i, evt.Id(), expectedEvent.Id())
			}
		}
		sch.EndRun()
	}
}

var emptyParams = []reflect.Value{}

var addRequestTests = []struct {
	requests []Request
	nodes    map[int]*MockNode
	events   []event.FunctionEvent
	err      bool
}{
	{
		[]Request{},
		map[int]*MockNode{},
		[]event.FunctionEvent{},
		true,
	},
	{
		[]Request{
			{0, "Foo", emptyParams},
			{1, "Foo", emptyParams},
			{0, "Foo", emptyParams},
		},
		map[int]*MockNode{0: {}, 1: {}},
		[]event.FunctionEvent{
			event.NewFunctionEvent(0, 0, "Foo", emptyParams...),
			event.NewFunctionEvent(1, 1, "Foo", emptyParams...),
			event.NewFunctionEvent(2, 0, "Foo", emptyParams...),
		},
		false,
	},
	{
		[]Request{
			{5, "Foo", []reflect.Value{}},
			{1, "Foo", []reflect.Value{}},
			{10, "Foo", []reflect.Value{}},
		},
		map[int]*MockNode{0: {}, 1: {}},
		[]event.FunctionEvent{
			event.NewFunctionEvent(0, 1, "Foo", emptyParams...),
		},
		false,
	},
	{
		[]Request{
			{5, "Foo", []reflect.Value{}},
			{1, "Foo", []reflect.Value{}},
			{10, "Foo", []reflect.Value{}},
		},
		map[int]*MockNode{0: {}, 3: {}},
		[]event.FunctionEvent{},
		true,
	},
}

func TestExecuteRun(t *testing.T) {
	for i, test := range executeRunTest {
		sch := NewMockRunScheduler(test.events...)
		gsm := NewMockStateManager()
		sm := stateManager.NewRunStateManager[MockNode, State](gsm, GetState)
		sim := newRunSimulator(
			sch,
			sm,
			1000,
			false,
		)

		err := sim.executeRun(test.nodes)
		isErr := (err != nil)
		if isErr != test.expectedErr {
			if isErr {
				t.Errorf("Test %v: Expected no error got: %v", i, err)
			} else {
				t.Errorf("Test %v: Expected to receive an error", i)
			}
			continue
		}

		sm.EndRun()

		actual := gsm.receivedRun
		expected := test.expectedStates

		if len(actual) != len(expected) {
			t.Errorf("Test %v: Unexpected length of state. Got: %v. Expected: %v", i, len(actual), len(expected))
		}

		for j := 0; j < len(gsm.receivedRun); j++ {
			if !maps.Equal(actual[j].LocalStates, expected[j]) {
				t.Errorf("Test %v: Unexpected state. Got: %v. Expected: %v. \n Full run: %v", i, actual[j].LocalStates, expected[j], actual)
			}
		}
	}
}

var executeRunTest = []struct {
	nodes          map[int]*MockNode
	events         []event.Event
	expectedErr    bool
	expectedStates []map[int]State
}{
	{
		// Execute 1 event
		map[int]*MockNode{0: {}, 1: {}, 2: {}},
		[]event.Event{
			MockEvent{0, 0, false, 1},
		},
		false,
		[]map[int]State{
			{0: {1}, 1: {0}, 2: {0}},
		},
	},
	{
		// Execute event on non-existing process
		map[int]*MockNode{0: {}, 1: {}, 2: {}},
		[]event.Event{
			MockEvent{0, 5, false, 1},
		},
		true,
		[]map[int]State{},
	},
	{
		// Execute 5 correct events on different nodes event
		map[int]*MockNode{0: {}, 1: {}, 2: {}},
		[]event.Event{
			MockEvent{0, 0, false, 1},
			MockEvent{1, 1, false, 1},
			MockEvent{2, 2, false, 1},
			MockEvent{3, 0, false, 3},
			MockEvent{4, 0, false, 5},
		},
		false,
		[]map[int]State{
			{0: {1}, 1: {0}, 2: {0}},
			{0: {1}, 1: {1}, 2: {0}},
			{0: {1}, 1: {1}, 2: {1}},
			{0: {3}, 1: {1}, 2: {1}},
			{0: {5}, 1: {1}, 2: {1}},
		},
	},
	{
		// Execute 5 correct events on different nodes event
		map[int]*MockNode{0: {}, 1: {}, 2: {}},
		[]event.Event{
			MockEvent{0, 0, false, 1},
			MockEvent{1, 1, false, 1},
			MockEvent{2, 2, false, 1},
			MockEvent{3, 4, false, 3},
			MockEvent{4, 0, false, 5},
		},
		true,
		[]map[int]State{
			{0: {1}, 1: {0}, 2: {0}},
			{0: {1}, 1: {1}, 2: {0}},
			{0: {1}, 1: {1}, 2: {1}},
		},
	},
}

func TestExecuteEventDontIgnorePanics(t *testing.T) {
	sch := NewMockRunScheduler()
	gsm := NewMockStateManager()
	sm := stateManager.NewRunStateManager[MockNode, State](gsm, GetState)
	sim := newRunSimulator(
		sch,
		sm,
		1000,
		false,
	)

	// The value -1 is hard coded to trigger a panic
	evt := MockEvent{0, 0, false, -1}
	n := &MockNode{}
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Did not expect code to panic")
		}
	}()

	err := sim.executeEvent(n, evt)
	if err == nil {
		t.Errorf("Expected to receive an error")
	}
}

func TestInitRun(t *testing.T) {
	for i, test := range initRunTest {
		sch := NewMockRunScheduler()
		gsm := NewMockStateManager()
		sm := stateManager.NewRunStateManager[MockNode, State](gsm, GetState)
		sim := newRunSimulator(
			sch,
			sm,
			1000,
			false,
		)

		sch.runEnded = test.runEnded

		nodes, err := sim.initRun(test.initNodes, test.failingNodes, func(t *MockNode) { t.crashed = true }, test.requests...)
		isErr := (err != nil)
		if isErr != test.expectedError {
			if isErr {
				t.Errorf("Test %v: Expected no error got: %v", i, err)
			} else {
				t.Errorf("Test %v: Expected to receive an error", i)
			}
			continue
		}

		if !reflect.DeepEqual(nodes, test.expectedNodes) {
			t.Errorf("Test %v: Unexpected node map returned. Got %v. Expected: %v", i, nodes, test.expectedNodes)
		}
		if !slices.EqualFunc(sch.addedEvents, test.expectedEvents, event.EventsEquals) {
			t.Errorf("Test %v: Unexpected events added. Got: %v, Expected: %v", i, sch.addedEvents, test.expectedEvents)
		}
	}
}

var initRunTest = []struct {
	initNodes    func(SimulationParameters) map[int]*MockNode
	failingNodes []int
	requests     []Request
	runEnded     bool

	expectedError  bool
	expectedNodes  map[int]*MockNode
	expectedEvents []event.Event
}{
	{
		func(sp SimulationParameters) map[int]*MockNode {
			return map[int]*MockNode{0: {}, 1: {}, 2: {}}
		},
		[]int{},
		[]Request{NewRequest(0, "Foo")},
		false,

		false,
		map[int]*MockNode{0: {}, 1: {}, 2: {}},
		[]event.Event{
			event.NewFunctionEvent(0, 0, "Foo"),
		},
	},
	{
		func(sp SimulationParameters) map[int]*MockNode {
			return map[int]*MockNode{0: {}, 1: {}, 2: {}}
		},
		[]int{},
		[]Request{},
		false,

		true,
		nil,
		[]event.Event{},
	},
	{
		func(sp SimulationParameters) map[int]*MockNode {
			return map[int]*MockNode{0: {}, 1: {val: 3}, 2: {val: 5}}
		},
		[]int{},
		[]Request{NewRequest(0, "Foo")},
		false,

		false,
		map[int]*MockNode{0: {}, 1: {val: 3}, 2: {val: 5}},
		[]event.Event{
			event.NewFunctionEvent(0, 0, "Foo"),
		},
	},
	{
		func(sp SimulationParameters) map[int]*MockNode {
			return map[int]*MockNode{0: {}, 5: {val: 3}, 3: {val: 5}}
		},
		[]int{},
		[]Request{NewRequest(0, "Foo")},
		false,

		false,
		map[int]*MockNode{0: {}, 5: {val: 3}, 3: {val: 5}},
		[]event.Event{
			event.NewFunctionEvent(0, 0, "Foo"),
		},
	},
	{
		func(sp SimulationParameters) map[int]*MockNode {
			return map[int]*MockNode{0: {}, 5: {val: 3}, 3: {val: 5}}
		},
		[]int{3, 5},
		[]Request{NewRequest(0, "Foo")},
		false,

		false,
		map[int]*MockNode{0: {}, 5: {val: 3}, 3: {val: 5}},
		[]event.Event{
			event.NewFunctionEvent(0, 0, "Foo"),
			event.NewCrashEvent(3, func(i int) error { return nil }, func() {}),
			event.NewCrashEvent(5, func(i int) error { return nil }, func() {}),
		},
	},
	{
		func(sp SimulationParameters) map[int]*MockNode {
			return map[int]*MockNode{0: {}, 5: {val: 3}, 3: {val: 5}}
		},
		[]int{3, 5, 10},
		[]Request{NewRequest(0, "Foo"), NewRequest(2, "foo")},
		false,

		false,
		map[int]*MockNode{0: {}, 5: {val: 3}, 3: {val: 5}},
		[]event.Event{
			event.NewFunctionEvent(0, 0, "Foo"),
			event.NewCrashEvent(3, func(i int) error { return nil }, func() {}),
			event.NewCrashEvent(5, func(i int) error { return nil }, func() {}),
		},
	},
	{
		func(sp SimulationParameters) map[int]*MockNode {
			return map[int]*MockNode{0: {}, 5: {val: 3}, 3: {val: 5}}
		},
		[]int{3, 5, 10},
		[]Request{NewRequest(0, "Foo"), NewRequest(2, "foo")},
		true,

		true,
		nil,
		[]event.Event{},
	},
	{
		func(sp SimulationParameters) map[int]*MockNode {
			return map[int]*MockNode{0: {}, 5: {val: 3}, 3: {val: 5}}
		},
		[]int{},
		[]Request{NewRequest(0, "Foo")},
		true,

		true,
		nil,
		[]event.Event{},
	},
}

func TestMainLoop(t *testing.T) {
	for i, test := range mainLoopTest {
		sch := NewMockGlobalScheduler()
		sim := NewSimulator[MockNode, State](sch, test.ignoreError, false, test.maxRuns, 1000, 10)

		nextRun := make(chan bool)
		status := make(chan error)
		closing := make(chan bool)

		numNextRuns := 0
		go func() {
			for range nextRun {
				if numNextRuns >= len(test.status) {
					break
				}

				status <- test.status[numNextRuns]
				numNextRuns++
			}
			closing <- true
		}()

		nextRun <- true

		err := sim.mainLoop(1, test.startedRuns, nextRun, status, closing)
		isErr := (err != nil)
		if isErr != test.expectedErr {
			if isErr {
				t.Errorf("Test %v: Expected no error got: %v", i, err)
			} else {
				t.Errorf("Test %v: Expected to receive an error", i)
			}
			continue
		}

		if numNextRuns != test.expectedNextRun {
			t.Errorf("Test %v: Unexpected number of performed runs. Got: %v. Expected: %v", i, numNextRuns, test.expectedNextRun)
		}
	}
}

var (
	noError error
	err     = fmt.Errorf("Dummy error")
)

var mainLoopTest = []struct {
	ignoreError bool
	maxRuns     int
	startedRuns int // Number of started runs before the simulation begins.

	status          []error
	expectedNextRun int
	expectedErr     bool
}{
	{
		false,
		100,
		1,
		[]error{noError, noError, noError, noError, noError},

		5,
		false,
	},
	{
		false,
		100,
		1,

		[]error{noError, noError, err},
		3,
		true,
	},
	{
		false,
		100,
		1,

		[]error{noError, noError, err, noError, noError},
		3, // Since ignoreError is false it will stop at the error
		true,
	},
	{
		false,
		100,
		99,

		[]error{noError, noError, err},
		2,
		false, // Will never get to the error, since maxRuns is reached first
	},
	{
		true,
		100,
		99,

		[]error{noError, noError, err},
		2,
		false, // Will never get to the error, since maxRuns is reached first
	},
	{
		true,
		15,
		10,

		[]error{noError, noError, err, noError, noError},
		5,
		true, // Will get to the error, but will continue executing since ignoreError is true
	},
}
