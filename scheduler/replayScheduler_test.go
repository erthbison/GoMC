package scheduler

import (
	"errors"
	"testing"

	"golang.org/x/exp/slices"
)

func TestReplayScheduler(t *testing.T) {
	for i, test := range replaySchedulerTest {
		sch := NewReplayScheduler(test.run)
		for _, evt := range test.events {
			sch.AddEvent(evt)
		}
		actualRun := []string{}
		for {
			evt, err := sch.GetEvent()
			if err == NoEventError {
				break
			}
			if err == RunEndedError {
				break
			}
			if evt == nil {
				if !test.expectedErr {
					t.Errorf("Did not expect to receive an error in test %v", i)
				}
				break
			}
			actualRun = append(actualRun, evt.Id())
		}
		if !test.expectedErr && !slices.Equal(actualRun, test.run) {
			t.Errorf("Received unexpected order of events in test %v. \n Got: %v\n Expected %v", i+1, actualRun, test.run)
		}
	}
}

func TestReplayNodeCrash(t *testing.T) {
	run := []string{"1", "3", "4", "6", "7"}
	events := []MockEvent{
		{id: 4, target: 0},
		{id: 5, target: 1},
		{id: 1, target: 0},
		{id: 3, target: 0},
		{id: 7, target: 0},
		{id: 2, target: 1},
		{id: 6, target: 0},
	}
	sch := NewReplayScheduler(run)
	for _, evt := range events {
		sch.AddEvent(evt)
	}
	sch.NodeCrash(1)
	actualRun := []string{}
	for {
		evt, err := sch.GetEvent()
		if errors.Is(err, NoEventError) {
			break
		}
		if errors.Is(err, RunEndedError) {
			break
		}
		actualRun = append(actualRun, evt.Id())
	}
	if !slices.Equal(actualRun, run) {
		t.Errorf("Received unexpected run. \nGot: %v. \nExpected: %v", actualRun, run)
	}
}

func TestReplaySchedulerEndRun(t *testing.T) {
	run := []string{"1", "3", "4", "6", "7"}
	events := []MockEvent{
		{id: 4, target: 0},
		{id: 5, target: 1},
		{id: 1, target: 0},
		{id: 3, target: 0},
		{id: 7, target: 0},
		{id: 2, target: 1},
		{id: 6, target: 0},
	}
	sch := NewReplayScheduler(run)
		for _, evt := range events {
			sch.AddEvent(evt)
		}
		actualRun := []string{}
		for {
			evt, err := sch.GetEvent()
			if errors.Is(err, NoEventError) {
				break
			}
			if errors.Is(err, RunEndedError) {
				break
			}
			actualRun = append(actualRun, evt.Id())
		}
		if !slices.Equal(actualRun, run) {
			t.Errorf("Received unexpected run. \nGot: %v. \nExpected: %v", actualRun, run)
		}
		sch.EndRun()

}

var replaySchedulerTest = []struct {
	run         []string
	events      []MockEvent
	expectedErr bool
}{
	{
		run:         []string{},
		events:      []MockEvent{},
		expectedErr: true,
	},
	{
		run:         []string{"1", "2"},
		events:      []MockEvent{{id: 2}, {id: 1}},
		expectedErr: false,
	},
	{
		// The provided run contains an event that is not added. Expect an error
		run:         []string{"1", "3"},
		events:      []MockEvent{{id: 2}, {id: 1}},
		expectedErr: true,
	},
	{
		run:         []string{"1", "2", "2"},
		events:      []MockEvent{{id: 2}, {id: 1}, {id: 3}},
		expectedErr: true,
	},
	{
		run:         []string{"1", "2", "2"},
		events:      []MockEvent{{id: 2}, {id: 1}, {id: 2}},
		expectedErr: false,
	},
}
