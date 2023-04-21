package scheduler

import (
	"gomc/event"
	"sync"
)

// A scheduler that initially will follow a provided run before it begin searching the state space.
type GuidedSearch struct {
	run    []event.EventId
	search GlobalScheduler
}

func NewGuidedSearch(search GlobalScheduler, run []event.EventId) *GuidedSearch {
	return &GuidedSearch{
		run:    run,
		search: search,
	}
}

func (gs *GuidedSearch) GetRunScheduler() RunScheduler {
	search := gs.search.GetRunScheduler()
	return newRunGuidedSearch(search, gs.run)
}

type runGuidedSearch struct {
	sync.Mutex
	// The scheduler used to search the state space
	search RunScheduler

	// The provided run
	run    []event.EventId
	guided *runReplay

	useGuided bool
}

// Create a new GuidedSearch scheduler using the provided search scheduler for searching the state space after it hasa followed the provided run
func newRunGuidedSearch(search RunScheduler, run []event.EventId) *runGuidedSearch {
	return &runGuidedSearch{
		search: search,

		run:       run,
		guided:    nil,
		useGuided: true,
	}
}

// Get the next event in the run. Will return RunEndedError if there are no more events in the run. Will return NoEventError if there are no more available events in any run.
// The event returned must be an event that has been added during the current run.
// Will follow the provided run until it has been completed or until it is unable to find the next event.
// After that the scheduler will begin to search the state space using teh provided search scheduler.
func (gs *runGuidedSearch) GetEvent() (event.Event, error) {
	gs.Lock()
	defer gs.Unlock()
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
func (gs *runGuidedSearch) AddEvent(evt event.Event) {
	gs.Lock()
	defer gs.Unlock()
	if gs.useGuided {
		gs.guided.AddEvent(evt)
	} else {
		gs.search.AddEvent(evt)
	}
}

func (gs *runGuidedSearch) StartRun() error {
	gs.Lock()
	defer gs.Unlock()
	gs.useGuided = true
	gs.guided = newRunReplay(gs.run)
	err := gs.guided.StartRun()
	if err != nil {
		return err
	}
	return gs.search.StartRun()
}

// Finish the current run and prepare for the next one
func (gs *runGuidedSearch) EndRun() {
	gs.search.EndRun()
	gs.guided.EndRun()
}
