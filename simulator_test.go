package gomc

import (
	"gomc/event"
	"reflect"
	"testing"
)

func TestSimulatorNoEvents(t *testing.T) {
	// Test
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
	sim := newRunSimulator(
		sch,
		&RunStateManager[MockNode, State]{},
		1000,
		false,
	)
	for i, test := range addRequestTests {
		sim.scheduleRequests(test.requests, test.nodes)

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
}{
	{
		[]Request{},
		map[int]*MockNode{},
		[]event.FunctionEvent{},
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
	},
	{
		[]Request{
			{5, "Foo", []reflect.Value{}},
			{1, "Foo", []reflect.Value{}},
			{10, "Foo", []reflect.Value{}},
		},
		map[int]*MockNode{0: {}, 3: {}},
		[]event.FunctionEvent{
		},
	},
}
