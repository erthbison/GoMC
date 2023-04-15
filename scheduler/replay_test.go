package scheduler

import (
	"errors"
	"testing"

	"golang.org/x/exp/slices"
)

func TestReplayScheduler(t *testing.T) {
	for i, test := range replaySchedulerTest {
		gsch := NewReplay(test.run)
		sch := gsch.GetRunScheduler()
		err := sch.StartRun()
		if err != nil {
			t.Errorf("Received unexpected error: %v", err)
		}
		for _, evt := range test.events {
			sch.AddEvent(evt)
		}
		actualRun := []uint64{}
		for {
			evt, err := sch.GetEvent()
			if err == NoRunsError {
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

func TestReplaySchedulerEndRun(t *testing.T) {
	run := []uint64{1, 3, 4, 6, 7}
	events := []MockEvent{
		{id: 4, target: 0},
		{id: 5, target: 1},
		{id: 1, target: 0},
		{id: 3, target: 0},
		{id: 7, target: 0},
		{id: 2, target: 1},
		{id: 6, target: 0},
	}
	gsch := NewReplay(run)
	sch := gsch.GetRunScheduler()
	err := sch.StartRun()
	if err != nil {
		t.Errorf("Received unexpected error: %v", err)
	}
	for _, evt := range events {
		sch.AddEvent(evt)
	}
	actualRun := []uint64{}
	for {
		evt, err := sch.GetEvent()
		if errors.Is(err, NoRunsError) {
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
	run         []uint64
	events      []MockEvent
	expectedErr bool
}{
	{
		run:         []uint64{},
		events:      []MockEvent{},
		expectedErr: true,
	},
	{
		run:         []uint64{1, 2},
		events:      []MockEvent{{id: 2}, {id: 1}},
		expectedErr: false,
	},
	{
		// The provided run contains an event that is not added. Expect an error
		run:         []uint64{1, 3},
		events:      []MockEvent{{id: 2}, {id: 1}},
		expectedErr: true,
	},
	{
		run:         []uint64{1, 2, 2},
		events:      []MockEvent{{id: 2}, {id: 1}, {id: 3}},
		expectedErr: true,
	},
	{
		run:         []uint64{1, 2, 2},
		events:      []MockEvent{{id: 2}, {id: 1}, {id: 2}},
		expectedErr: false,
	},
}
