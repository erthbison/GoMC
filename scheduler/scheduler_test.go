package scheduler

import (
	"errors"
	"gomc/event"
	"strconv"
	"testing"
)

func TestBasicSchedulerNoEvents(t *testing.T) {
	sch := NewBasicScheduler()
	_, err := sch.GetEvent()
	if !errors.Is(err, RunEndedError) {
		t.Fatalf("unexpected error. Got %v. Expected: %v", err, RunEndedError)
	}
	_, err = sch.GetEvent()
	if !errors.Is(err, RunEndedError) {
		t.Fatalf("unexpected error. Got %v. Expected: %v", err, RunEndedError)
	}
	sch.EndRun()
}

func TestBasicSchedulerExplore2Events(t *testing.T) {
	sch := NewBasicScheduler()
	testDeterministicExplore2Events(t, sch)
}

func TestBasicExploreBranchingEvents(t *testing.T) {
	sch := NewBasicScheduler()
	testDeterministicExploreBranchingEvents(t, sch)
}

func TestBasicSchedulerNodeCrash(t *testing.T) {
	sch := NewBasicScheduler()
	testSchedulerCrash(sch, t)
}

func TestRandomScheduler(t *testing.T) {
	// Perform one random run
	sch := NewRandomScheduler(1)

	sch.AddEvent(MockEvent{0, 0, false})
	sch.AddEvent(MockEvent{1, 0, false})

	// This should cause two possible interleavings. Either event 1 first and Event 2 afterwards or Event 2 then Event 1.
	run := []event.Event{}
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
	sch.AddEvent(MockEvent{0, 0, false})
	sch.AddEvent(MockEvent{1, 0, false})

	_, err = sch.GetEvent()
	if !errors.Is(err, NoEventError) {
		t.Errorf("Expected to get a NoEventError. Got: %v", err)
	}
}

func TestRandomSchedulerNodeCrash(t *testing.T) {
	sch := NewRandomScheduler(2)
	testSchedulerCrash(sch, t)
}

func TestQueueSchedulerCrash(t *testing.T) {
	sch := NewQueueScheduler()
	testSchedulerCrash(sch, t)
}

func TestQueueSchedulerExplore2Events(t *testing.T) {
	sch := NewQueueScheduler()
	testDeterministicExplore2Events(t, sch)
}

func TestQueueExploreBranchingEvents(t *testing.T) {
	sch := NewQueueScheduler()
	testDeterministicExploreBranchingEvents(t, sch)
}


func BenchmarkBasicScheduler(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sch := NewBasicScheduler()
		err := benchmarkScheduler(sch, 7)
		if err != nil {
			b.Errorf(err.Error())
		}
	}
}

func BenchmarkQueueScheduler(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sch := NewQueueScheduler()
		err := benchmarkScheduler(sch, 7)
		if err != nil {
			b.Errorf(err.Error())
		}
	}
}

func benchmarkScheduler(sch Scheduler, numEvents int) error {
	for {
		evt, err := sch.GetEvent()
		if errors.Is(err, RunEndedError) {
			sch.EndRun()
			for i := 0; i < numEvents; i++ {
				sch.AddEvent(MockEvent{i, 0, false})
			}
		} else if errors.Is(err, NoEventError) {
			return nil
		}
		if err == nil && evt == nil {
			return errors.New("Expected to receive an event.")
		}
	}
}

func testDeterministicExplore2Events(t *testing.T, sch Scheduler) {
	sch.AddEvent(MockEvent{0, 0, false})
	sch.AddEvent(MockEvent{1, 0, false})

	// This should cause two possible interleavings. Either event 1 first and Event 2 afterwards or Event 2 then Event 1.
	run1 := []event.Event{}
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
	sch.AddEvent(MockEvent{0, 0, false})
	sch.AddEvent(MockEvent{1, 0, false})

	run2 := []event.Event{}
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

	sch.AddEvent(MockEvent{0, 0, false})
	sch.AddEvent(MockEvent{1, 0, false})
	_, err = sch.GetEvent()
	if !errors.Is(err, NoEventError) {
		t.Errorf("Expected to get a NoEventError. Got: %v", err)
	}
}

func testDeterministicExploreBranchingEvents(t *testing.T, sch Scheduler) {
	// 1. Add one event
	// 2. Get One event
	// 3. Add Two events
	// 4. Get the two events
	// Expects two chains. Both should contain all 3 events and start with event 0
	sch.AddEvent(MockEvent{0, 0, false})

	run1 := []event.Event{}
	evt, err := sch.GetEvent()
	run1 = append(run1, evt)
	if err != nil {
		t.Errorf("Expected no error. Got: %v", err)
	}
	if evt.Id() != "0" {
		t.Errorf("Expected to be returned Event 0. Got: %v", evt)
	}

	sch.AddEvent(MockEvent{1, 0, false})
	sch.AddEvent(MockEvent{2, 0, false})

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
	sch.AddEvent(MockEvent{0, 0, false})

	run2 := []event.Event{}
	evt, err = sch.GetEvent()
	run2 = append(run2, evt)
	if err != nil {
		t.Errorf("Expected no error. Got: %v", err)
	}
	if evt.Id() != "0" {
		t.Errorf("Expected to be returned Event 0. Got: %v", evt)
	}
	sch.AddEvent(MockEvent{1, 0, false})
	sch.AddEvent(MockEvent{2, 0, false})

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
	sch.AddEvent(MockEvent{0, 0, false})
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

func testSchedulerCrash(sch Scheduler, t *testing.T) {
	sch.AddEvent(MockEvent{0, 0, false})
	sch.AddEvent(MockEvent{1, 1, false})
	sch.AddEvent(MockEvent{2, 2, false})

	_, err := sch.GetEvent()
	if err != nil {
		t.Errorf("Expected no error. Got: %v", err)
	}

	sch.NodeCrash(1)

	evt, err := sch.GetEvent()
	if err != nil {
		t.Errorf("Expected no error. Got: %v", err)
	}
	if evt.Id() == "1" {
		t.Errorf("Event 1 should have been removed, but received it ")
	}

	sch.AddEvent(MockEvent{3, 1, false})

	evt, err = sch.GetEvent()
	if evt != nil {
		if evt.Id() == "3" {
			t.Errorf("Event 3 is targeting a disabled node so we should not receive it.")
		} else {
			t.Errorf("Expected to receive no event")
		}
	}
	if !errors.Is(err, RunEndedError) {
		t.Errorf("Expected %v, got: %v", RunEndedError, err)
	}
	sch.EndRun()

	sch.AddEvent(MockEvent{1, 1, false})
	_, err = sch.GetEvent()
	if err != nil {
		t.Errorf("Expected no error. Got: %v", err)
	}
	if err != nil {
		t.Errorf("Added an event to a node that had crashed in a previous run. Expected to receive the event but got an error.")
	}
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
