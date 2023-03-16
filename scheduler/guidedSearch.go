package scheduler

import (
	"gomc/event"
)

// A scheduler that initially will follow a provided run before it begin searching the state space.
type guidedSearch struct {
	// The scheduler used to search the state space
	search Scheduler

	// The provided run
	run    []string
	guided *replayScheduler

	useGuided bool
}

// Create a new GuidedSearch scheduler using the provided search scheduler for searching the state space after it hasa followed the provided run
func NewGuidedSearch(search Scheduler, run []string) *guidedSearch {
	guided := NewReplayScheduler(run)
	return &guidedSearch{
		search: search,

		run:       run,
		guided:    guided,
		useGuided: true,
	}
}

// Get the next event in the run. Will return RunEndedError if there are no more events in the run. Will return NoEventError if there are no more available events in any run.
// The event returned must be an event that has been added during the current run.
// Will follow the provided run until it has been completed or until it is unable to find the next event.
// After that the scheduler will begin to search the state space using teh provided search scheduler.
func (gs *guidedSearch) GetEvent() (event.Event, error) {
	if gs.useGuided {
		evt, err := gs.guided.GetEvent()
		if err != nil {
			gs.useGuided = false
			// Migrate all pending events from the guided scheduler to the search scheduler
			for _, evt := range gs.guided.pendingEvents {
				gs.search.AddEvent(evt)
			}
			return gs.search.GetEvent()
		}
		return evt, err
	} else {
		return gs.search.GetEvent()
	}
}

// Add an event to the list of possible events
func (gs *guidedSearch) AddEvent(evt event.Event) {
	if gs.useGuided {
		gs.guided.AddEvent(evt)
	} else {
		gs.search.AddEvent(evt)
	}
}

// Finish the current run and prepare for the next one
func (gs *guidedSearch) EndRun() {
	gs.useGuided = true
	gs.guided = NewReplayScheduler(gs.run)
	gs.search.EndRun()
}