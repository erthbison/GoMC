package scheduler

import (
	"errors"
	"testing"

	"golang.org/x/exp/slices"
)

func TestGuidedSearch(t *testing.T) {
	for i, test := range GuidedSearchTests {
		gsch := NewGuidedSearch(NewPrefix(), test.run)

		sch := gsch.GetRunScheduler()

		err := sch.StartRun()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		numRuns := 0
		actualRun := []string{}
		numEvent := 0
		for {
			if events, ok := test.events[numEvent]; ok {
				// add some more events
				for _, event := range events {
					sch.AddEvent(event)
				}
			}
			evt, err := sch.GetEvent()
			if errors.Is(err, RunEndedError) {
				sch.EndRun()
				numRuns++
				if !slices.Equal(test.run[:test.lengthReplay], actualRun[:test.lengthReplay]) {
					t.Errorf("Did not follow run for start of execution in test %v", i)
				}

				err := sch.StartRun()
				if errors.Is(err, NoRunsError) {
					break
				}
				actualRun = []string{}
				numEvent = 0
				continue
			}
			actualRun = append(actualRun, evt.Id())
			numEvent++
		}
		if numRuns != test.numRuns {
			t.Errorf("Got an unexpected number of total runs in test %v. Got %v. Expected: %v", i, numRuns, test.numRuns)
		}
	}
}

var GuidedSearchTests = []struct {
	run    []string
	events map[int][]MockEvent
	// Specify the number of events in the provided run that the scheduler is expected to follow before it begins searching
	// This is generally the length of the run, but can be shorter if for instance the replayScheduler is unable to find the next event in the run
	lengthReplay int
	numRuns      int
}{
	{
		run: []string{"1", "2", "3"},
		events: map[int][]MockEvent{
			0: {{id: 1}, {id: 2}, {id: 3}, {id: 4}, {id: 5}},
		},
		lengthReplay: 3,
		numRuns:      2,
	},
	{
		run: []string{},
		events: map[int][]MockEvent{
			0: {{id: 1}, {id: 2}, {id: 3}, {id: 4}, {id: 5}},
		},
		lengthReplay: 0,
		numRuns:      120,
	},
	{
		run: []string{"1", "2", "3", "4", "5"},
		events: map[int][]MockEvent{
			0: {{id: 1}, {id: 2}, {id: 3}, {id: 4}, {id: 5}},
		},
		lengthReplay: 5,
		numRuns:      1,
	},
	{
		run:          []string{"1", "2", "3", "4", "5"},
		events:       map[int][]MockEvent{0: {{id: 1}, {id: 2}, {id: 4}, {id: 5}}},
		lengthReplay: 2,
		numRuns:      2,
	},
	{
		run: []string{"1", "2"},
		events: map[int][]MockEvent{
			0: {{id: 1}, {id: 2}, {id: 3}, {id: 4}, {id: 5}},
			2: {{id: 6}, {id: 7}},
		},
		lengthReplay: 2,
		numRuns:      120,
	},
}
