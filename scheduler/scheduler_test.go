package scheduler

import (
	"errors"
	"gomc/event"
	"strconv"
	"testing"
)

func benchmarkRunScheduler(sch RunScheduler, numEvents int) error {
	for {
		evt, err := sch.GetEvent()
		if errors.Is(err, RunEndedError) {
			sch.EndRun()
			for i := 0; i < numEvents; i++ {
				sch.AddEvent(MockEvent{event.EventId(strconv.Itoa(i)), 0, false})
			}
		} else if errors.Is(err, NoRunsError) {
			return nil
		}
		if err == nil && evt == nil {
			return errors.New("Expected to receive an event.")
		}
	}
}

func testDeterministicExplore2Events(t *testing.T, gsch GlobalScheduler) {
	sch := gsch.GetRunScheduler()
	err := sch.StartRun()
	if err != nil {
		t.Errorf("Did not expect to receive an error. Got %v", err)
	}
	sch.AddEvent(MockEvent{"0", 0, false})
	sch.AddEvent(MockEvent{"1", 0, false})

	// This should cause two possible interleavings. Either event 1 first and Event 2 afterwards or Event 2 then Event 1.
	run1 := []event.Event{}
	for i := 0; i < 2; i++ {
		evt, err := sch.GetEvent()
		if err != nil {
			t.Errorf("Did not expect to receive an error. Got %v", err)
		}
		run1 = append(run1, evt)
	}
	_, err = sch.GetEvent()
	if !errors.Is(err, RunEndedError) {
		t.Errorf("Expected to get a RunEndedError. Got: %v", err)
	}
	sch.EndRun()

	err = sch.StartRun()
	if err != nil {
		t.Errorf("Did not expect to receive an error. Got %v", err)
	}
	sch.AddEvent(MockEvent{"0", 0, false})
	sch.AddEvent(MockEvent{"1", 0, false})

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

	err = sch.StartRun()
	if !errors.Is(err, NoRunsError) {
		t.Errorf("Expected to get a NoEventError. Got: %v", err)
	}
}

func testDeterministicExploreBranchingEvents(t *testing.T, gsch GlobalScheduler) {
	// 1. Add one event
	// 2. Get One event
	// 3. Add Two events
	// 4. Get the two events
	// Expects two chains. Both should contain all 3 events and start with event 0
	sch := gsch.GetRunScheduler()

	err := sch.StartRun()
	if err != nil {
		t.Errorf("Did not expect to receive an error. Got %v", err)
	}
	sch.AddEvent(MockEvent{"0", 0, false})

	run1 := []event.Event{}
	evt, err := sch.GetEvent()
	run1 = append(run1, evt)
	if err != nil {
		t.Errorf("Expected no error. Got: %v", err)
	}
	if evt.Id() != "0" {
		t.Errorf("Expected to be returned Event 0. Got: %v", evt)
	}

	sch.AddEvent(MockEvent{"1", 0, false})
	sch.AddEvent(MockEvent{"2", 0, false})

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

	err = sch.StartRun()
	if err != nil {
		t.Errorf("Did not expect to receive an error. Got %v", err)
	}
	sch.AddEvent(MockEvent{"0", 0, false})

	run2 := []event.Event{}
	evt, err = sch.GetEvent()
	run2 = append(run2, evt)
	if err != nil {
		t.Errorf("Expected no error. Got: %v", err)
	}
	if evt.Id() != "0" {
		t.Errorf("Expected to be returned Event 0. Got: %v", evt)
	}
	sch.AddEvent(MockEvent{"1", 0, false})
	sch.AddEvent(MockEvent{"2", 0, false})

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

	err = sch.StartRun()
	if !errors.Is(err, NoRunsError) {
		t.Errorf("Expected to get a NoEventError. Got: %v", err)
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

func testConcurrentDeterministic(t *testing.T, gsch GlobalScheduler) {
	runs := [][]event.Event{}
	runChan := make(chan []event.Event)
	for i := 0; i < 2; i++ {
		go func() {
			sch := gsch.GetRunScheduler()
			for {
				err := sch.StartRun()
				if errors.Is(err, NoRunsError) {
					break
				}
				sch.AddEvent(MockEvent{"0", 0, false})

				run := []event.Event{}
				evt, err := sch.GetEvent()
				run = append(run, evt)
				if err != nil {
					t.Errorf("Expected no error. Got: %v", err)
				}
				if evt.Id() != "0" {
					t.Errorf("Expected to be returned Event 0. Got: %v", evt)
				}

				sch.AddEvent(MockEvent{"1", 0, false})
				sch.AddEvent(MockEvent{"2", 0, false})

				for i := 0; i < 2; i++ {
					evt, err = sch.GetEvent()
					if err != nil {
						t.Errorf("Expected no error. Got: %v", err)
					}
					run = append(run, evt)
				}
				_, err = sch.GetEvent()
				if !errors.Is(err, RunEndedError) {
					t.Errorf("Expected %v, got: %v", RunEndedError, err)
				}
				sch.EndRun()
				runChan <- run
			}
		}()
	}
	for run := range runChan {
		runs = append(runs, run)
		if len(runs) == 2 {
			close(runChan)
		}
	}

	if runs[0][0].Id() != runs[1][0].Id() {
		t.Errorf("Both chains should start with the same event. Chain1 %v, Chain2: %v", runs[0][0], runs[1][0])
	}
	for i := 1; i < 3; i++ {
		if runs[0][i].Id() != runs[1][3-i].Id() {
			t.Errorf("Unexpected result from the two runs. Expected run 2 to be reverse of run 1. Got: Run1: %v, Run2: %v", runs[0], runs[1])
		}
	}
}

type MockEvent struct {
	id       event.EventId
	target   int
	executed bool
}

func (me MockEvent) Id() event.EventId {
	return me.id
}

func (me MockEvent) Execute(_ any, chn chan error) {
	chn <- nil
}

func (me MockEvent) Target() int {
	return me.target
}
