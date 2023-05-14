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

// Create a new GuidedSearch scheduled
//
// search is the Scheduler that will be used to explore the state space after the provided run has been completed.
// run is the sequence of events that will be executed before beginning to search the state space.
func NewGuidedSearch(search GlobalScheduler, run []event.EventId) *GuidedSearch {
	return &GuidedSearch{
		run:    run,
		search: search,
	}
}

// Create a RunScheduler that will communicate with the global scheduler
func (gs *GuidedSearch) GetRunScheduler() RunScheduler {
	search := gs.search.GetRunScheduler()
	return newRunGuidedSearch(search, gs.run)
}

// Reset the global state of the GlobalScheduler.
// Prepare the scheduler for the next simulation.
func (gs *GuidedSearch) Reset() {
	gs.search.Reset()
}

// Manages the exploration of the state space in a single goroutine.
// Events can safely be added from multiple goroutines.
// Events will only be retrieved from a single goroutine during the simulation.
// Communicates with the GlobalScheduler to ensure that the state exploration remains consistent.
type runGuidedSearch struct {
	sync.Mutex
	// The scheduler used to search the state space
	search RunScheduler

	// The provided run
	run    []event.EventId
	guided *runReplay

	useGuided bool
}

// Create a new GuidedSearch scheduler using the provided search scheduler for searching the state space after it has followed the provided run
func newRunGuidedSearch(search RunScheduler, run []event.EventId) *runGuidedSearch {
	return &runGuidedSearch{
		search: search,

		run:       run,
		guided:    nil,
		useGuided: true,
	}
}

// Get the next event in the run.
//
// Will follow the provided run until it has been completed or until it is unable to find the next event.
// After that the scheduler will begin to search the state space using teh provided search scheduler.
//
// Will return RunEndedError if there are no more events in the run.
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

// Implements the event adder interface.
//
// It must be safe to add events from different goroutines.
// StartRun, EndRun and GetEvent will always be called from the same goroutine,
// but not from the same goroutine as AddEvent.
func (gs *runGuidedSearch) AddEvent(evt event.Event) {
	gs.Lock()
	defer gs.Unlock()
	if gs.useGuided {
		gs.guided.AddEvent(evt)
	} else {
		gs.search.AddEvent(evt)
	}
}

// Prepare for starting a new run.
//
// Returns a NoRunsError if all possible runs have been completed.
// May block until new runs are available.
// StartRun, EndRun and GetEvent will always be called from the same goroutine,
// but not from the same goroutine as AddEvent.
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

// Finish the current run and prepare for the next one.
//
// Will always be called after a run has been completely executed,
// even if an error occurred during execution of the run.
// StartRun, EndRun and GetEvent will always be called from the same goroutine,
// but not from the same goroutine as AddEvent.
func (gs *runGuidedSearch) EndRun() {
	gs.search.EndRun()
	gs.guided.EndRun()
}
