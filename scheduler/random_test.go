package scheduler

import (
	"errors"
	"gomc/event"
	"testing"
)

func TestRandomScheduler(t *testing.T) {
	// Perform one random run
	gsch := NewRandom(1)
	sch := gsch.GetRunScheduler()
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
	if !errors.Is(err, NoRunsError) {
		t.Errorf("Expected to get a NoEventError. Got: %v", err)
	}
}
