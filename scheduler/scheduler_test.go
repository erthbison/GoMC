package scheduler

import (
	"errors"
	"gomc/event"
	"strconv"
	"testing"
)

type node struct{}

func (n *node) Foo(from, to int, msg []byte) {}
func (n *node) Bar(from, to int, msg []byte) {}

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

func TestBasicSchedulerNoEvents(t *testing.T) {
	sch := NewBasicScheduler[node]()
	_, err := sch.GetEvent()
	if !errors.Is(err, RunEndedError) {
		t.Fatalf("unexpected error. Got %v. Expected: %v", err, RunEndedError)
	}
}

func TestBasicSchedulerExplore2Events(t *testing.T) {
	sch := NewBasicScheduler[node]()
	sch.AddEvent(MockEvent{0, false})
	sch.AddEvent(MockEvent{1, false})

	// This should cause two possible interleavings. Either event 1 first and Event 2 afterwards or Event 2 then Event 1.
	run1 := []event.Event[node]{}
	for i := 0; i < 2; i++ {
		evt, err := sch.GetEvent()
		if err != nil {
			t.Errorf("Did not expect to receive an error. Got %v", err)
		}
		run1 = append(run1, evt)
	}
	_, err := sch.GetEvent()
	if !errors.Is(err, RunEndedError) {
		t.Errorf("Expected to get a RunEndedError. Got: %v", err)
	}
	sch.EndRun()
	sch.AddEvent(MockEvent{0, false})
	sch.AddEvent(MockEvent{1, false})

	run2 := []event.Event[node]{}
	for i := 0; i < 2; i++ {
		evt, err := sch.GetEvent()
		if err != nil {
			t.Errorf("Did not expect to receive an error. Got %v", err)
		}
		run2 = append(run2, evt)
	}
	for i := 0; i < 2; i++ {
		if run1[i].Id() != run2[1-i].Id() {
			t.Errorf("Unexpected result from the two runs. Expected run 2 to be reverse of run 1. Got: Run1: %v, Run2: %v", run1, run2)
		}
	}
	_, err = sch.GetEvent()
	if !errors.Is(err, RunEndedError) {
		t.Errorf("Expected to get a RunEndedError. Got: %v", err)
	}
	sch.EndRun()

	sch.AddEvent(MockEvent{0, false})
	sch.AddEvent(MockEvent{1, false})
	_, err = sch.GetEvent()
	if !errors.Is(err, NoEventError) {
		t.Errorf("Expected to get a NoEventError. Got: %v", err)
	}

}

func TestBasicSchedulerExploreBranchingEvents(t *testing.T) {
	// 1. Add one event
	// 2. Get One event
	// 3. Add Two events
	// 4. Get the two events
	// Expects two chains. Both should contain all 3 events and start with event 0
	sch := NewBasicScheduler[node]()
	sch.AddEvent(MockEvent{0, false})

	run1 := []event.Event[node]{}
	evt, err := sch.GetEvent()
	run1 = append(run1, evt)
	if err != nil {
		t.Errorf("Expected no error. Got: %v", err)
	}
	if evt.Id() != "0" {
		t.Errorf("Expected to be returned Event 0. Got: %v", evt)
	}

	sch.AddEvent(MockEvent{1, false})
	sch.AddEvent(MockEvent{2, false})

	for i := 0; i < 2; i++ {
		evt, err = sch.GetEvent()
		if err != nil {
			t.Errorf("Expected no error. Got: %v", err)
		}
		run1 = append(run1, evt)
	}
	_, err = sch.GetEvent()
	if !errors.Is(err, RunEndedError) {
		t.Errorf("Expected %v, got: %v", RunEndedError, err)
	}
	sch.EndRun()
	sch.AddEvent(MockEvent{0, false})

	run2 := []event.Event[node]{}
	evt, err = sch.GetEvent()
	run2 = append(run2, evt)
	if err != nil {
		t.Errorf("Expected no error. Got: %v", err)
	}
	if evt.Id() != "0" {
		t.Errorf("Expected to be returned Event 0. Got: %v", evt)
	}
	sch.AddEvent(MockEvent{1, false})
	sch.AddEvent(MockEvent{2, false})

	for i := 0; i < 2; i++ {
		evt, err = sch.GetEvent()
		if err != nil {
			t.Errorf("Expected no error. Got: %v", err)
		}
		run2 = append(run2, evt)
	}
	_, err = sch.GetEvent()
	if !errors.Is(err, RunEndedError) {
		t.Errorf("Expected %v, got: %v", RunEndedError, err)
	}
	sch.EndRun()
	sch.AddEvent(MockEvent{0, false})
	_, err = sch.GetEvent()
	if !errors.Is(err, NoEventError) {
		t.Errorf("Expected %v, got: %v", NoEventError, err)
	}

	if run1[0].Id() != run2[0].Id() {
		t.Errorf("Both chains should start with the same event. Chain1 %v, Chain2: %v", run1[0], run2[0])
	}
	for i := 1; i < 3; i++ {
		if run1[i].Id() != run2[3-i].Id() {
			t.Errorf("Unexpected result from the two runs. Expected run 2 to be reverse of run 1. Got: Run1: %v, Run2: %v", run1, run2)
		}
	}

}

func TestRandomScheduler(t *testing.T) {
	// Perform one random run
	sch := NewRandomScheduler[node](1)

	sch.AddEvent(MockEvent{0, false})
	sch.AddEvent(MockEvent{1, false})

	// This should cause two possible interleavings. Either event 1 first and Event 2 afterwards or Event 2 then Event 1.
	run := []event.Event[node]{}
	for i := 0; i < 2; i++ {
		evt, err := sch.GetEvent()
		if err != nil {
			t.Errorf("Did not expect to receive an error. Got %v", err)
		}
		run = append(run, evt)
	}
	_, err := sch.GetEvent()
	if !errors.Is(err, RunEndedError) {
		t.Errorf("Expected to get a RunEndedError. Got: %v", err)
	}

	events := map[string]int{"0": 0, "1": 0}
	for _, evt := range run {
		if events[evt.Id()] > 1 {
			t.Errorf("Event occurred more times than it was scheduled: %v", evt.Id())
			events[evt.Id()]++
		}
	}
	sch.EndRun()
	sch.AddEvent(MockEvent{0, false})
	sch.AddEvent(MockEvent{1, false})

	_, err = sch.GetEvent()
	if !errors.Is(err, NoEventError) {
		t.Errorf("Expected to get a NoEventError. Got: %v", err)
	}
}
