package scheduler

import (
	"errors"
	"gomc/event"
	"sync"
)

// A scheduler that replays the provided run
//
// The scheduler replays the run once, before stopping the simulation.
// It does not further explore the state space.
//
// If the algorithm has been changed, and the scheduler is unable to follow the provided run it will return errors.
// It will stop after replaying the provided run, even if there are more pending events.
type Replay struct {
	run  []event.EventId
	done bool
}

// Create a new Replay scheduler
//
// run is the sequence of events that will be replayed.
func NewReplay(run []event.EventId) *Replay {
	return &Replay{
		run: run,
	}
}

// Create a RunScheduler that will communicate with the global scheduler
//
// The first run-specific scheduler will replay the run once. The other will immediately return NoRunsErrors.
func (r *Replay) GetRunScheduler() RunScheduler {
	if r.done {
		return newRunReplay(nil)
	}
	r.done = true
	return newRunReplay(r.run)
}

// Reset the global state of the GlobalScheduler.
// Prepare the scheduler for the next simulation.
func (r *Replay) Reset() {
	r.done = false
}

// Manages the exploration of the state space in a single goroutine.
// Events can safely be added from multiple goroutines.
// Events will only be retrieved from a single goroutine during the simulation.
// Communicates with the GlobalScheduler to ensure that the state exploration remains consistent.
type runReplay struct {
	sync.Mutex
	// A slice of the run to be replayed with event ids in order
	run []event.EventId
	// The index of the current event
	index int

	pendingEvents []event.Event
}

// Create a new runReplay scheduler
func newRunReplay(run []event.EventId) *runReplay {
	return &runReplay{
		index: 0,
		run:   run,

		pendingEvents: make([]event.Event, 0),
	}
}

// Get the next event in the run.
//
// Gets the next event in the provided run.
// Returns an error if it is unable to find the event.
//
// Will return RunEndedError if there are no more events in the run.
// The event returned must be an event that has been added during the current run.
// StartRun, EndRun and GetEvent will always be called from the same goroutine,
// but not from the same goroutine as AddEvent.
func (rr *runReplay) GetEvent() (event.Event, error) {
	rr.Lock()
	defer rr.Unlock()

	if rr.index >= len(rr.run) {
		return nil, RunEndedError
	}
	evtId := rr.run[rr.index]
	evt := rr.popEvent(evtId)
	if evt == nil {
		return nil, errors.New("RunScheduler: Unable to find next event")
	}
	rr.index++
	return evt, nil
}

// Get the event from the pending events
// Return nil if it is not found in the pending events
func (rr *runReplay) popEvent(id event.EventId) event.Event {
	for i, evt := range rr.pendingEvents {
		if evt.Id() == id {
			// Remove the event from the pending events and return the event
			rr.pendingEvents = append(rr.pendingEvents[:i], rr.pendingEvents[i+1:]...)
			return evt
		}
	}
	return nil
}

// Add an event to the list of possible events
//
// It must be safe to add events from different goroutines.
// StartRun, EndRun and GetEvent will always be called from the same goroutine,
// but not from the same goroutine as AddEvent.
func (rr *runReplay) AddEvent(evt event.Event) {
	rr.Lock()
	defer rr.Unlock()
	rr.pendingEvents = append(rr.pendingEvents, evt)
}

	// Prepare for starting a new run.
	//
	// Returns a NoRunsError if all possible runs have been completed.
	// May block until new runs are available.
	// StartRun, EndRun and GetEvent will always be called from the same goroutine,
	// but not from the same goroutine as AddEvent.
func (rr *runReplay) StartRun() error {
	rr.Lock()
	defer rr.Unlock()
	if rr.run == nil {
		return NoRunsError
	}
	return nil
}

// Finish the current run and prepare for the next one.
//
// Will always be called after a run has been completely executed,
// even if an error occurred during execution of the run.
// StartRun, EndRun and GetEvent will always be called from the same goroutine,
// but not from the same goroutine as AddEvent.
func (rr *runReplay) EndRun() {
	rr.Lock()
	defer rr.Unlock()
	rr.index = 0
	rr.run = nil
}
