package gomc_test

import (
	"errors"
	"gomc"
	"testing"
)

func TestBasicSchedulerNoEvents(t *testing.T) {
	sch := gomc.NewBasicScheduler[node]()
	_, err := sch.GetEvent()
	if !errors.Is(err, gomc.RunEndedError) {
		t.Fatalf("unexpected error. Got %v. Expected: %v", err, gomc.RunEndedError)
	}
}

func TestBasicSchedulerExplore2Events(t *testing.T) {
	sch := gomc.NewBasicScheduler[node]()
	sch.AddEvent(MockEvent{0, false})
	sch.AddEvent(MockEvent{1, false})

	// This should cause two possible interleavings. Either event 1 first and Event 2 afterwards or Event 2 then Event 1.
	run1 := []gomc.Event[node]{}
	for i := 0; i < 2; i++ {
		evt, err := sch.GetEvent()
		if err != nil {
			t.Errorf("Did not expect to receive an error. Got %v", err)
		}
		run1 = append(run1, evt)
	}
	_, err := sch.GetEvent()
	if !errors.Is(err, gomc.RunEndedError) {
		t.Errorf("Expected to get a RunEndedError. Got: %v", err)
	}
	sch.EndRun()
	sch.AddEvent(MockEvent{0, false})
	sch.AddEvent(MockEvent{1, false})

	run2 := []gomc.Event[node]{}
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
	if !errors.Is(err, gomc.RunEndedError) {
		t.Errorf("Expected to get a RunEndedError. Got: %v", err)
	}
	sch.EndRun()

	sch.AddEvent(MockEvent{0, false})
	sch.AddEvent(MockEvent{1, false})
	_, err = sch.GetEvent()
	if !errors.Is(err, gomc.NoEventError) {
		t.Errorf("Expected to get a NoEventError. Got: %v", err)
	}

}

func TestBasicSchedulerExploreBranchingEvents(t *testing.T) {
	// 1. Add one event
	// 2. Get One event
	// 3. Add Two events
	// 4. Get the two events
	// Expects two chains. Both should contain all 3 events and start with event 0
	sch := gomc.NewBasicScheduler[node]()
	sch.AddEvent(MockEvent{0, false})

	run1 := []gomc.Event[node]{}
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
	if !errors.Is(err, gomc.RunEndedError) {
		t.Errorf("Expected %v, got: %v", gomc.RunEndedError, err)
	}
	sch.EndRun()
	sch.AddEvent(MockEvent{0, false})

	run2 := []gomc.Event[node]{}
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
	if !errors.Is(err, gomc.RunEndedError) {
		t.Errorf("Expected %v, got: %v", gomc.RunEndedError, err)
	}
	sch.EndRun()
	sch.AddEvent(MockEvent{0, false})
	_, err = sch.GetEvent()
	if !errors.Is(err, gomc.NoEventError) {
		t.Errorf("Expected %v, got: %v", gomc.NoEventError, err)
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
