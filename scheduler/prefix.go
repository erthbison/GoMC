package scheduler

import (
	"errors"
	"gomc/event"
	"sync"
)

type run []event.EventId

// Explores the state space by maintaining a stack of unexplored prefixes.
// When a new run is started it follows the prefix and begins exploring from there, adding new prefixes it discovers as it executes events.
//
// Deterministic, stateful scheduler that explores the entire state space.
type Prefix struct {
	// unexplored prefixes
	r []run

	// Used to wait for a change in p.ongoing or p.r.
	// The condition is len(p.r) == 0 and p.ongoing > 0
	cond *sync.Cond

	// Number of runScheduler currently scheduling a run.
	// I.e. number of runScheduler not waiting for a new run
	ongoing int
}

// Create a Prefix Scheduler
//
// The Prefix Scheduler is a deterministic and stateful scheduler that explores the entire state space.
// Given enough runs it will completely explore the state space.
func NewPrefix() *Prefix {
	ss := &Prefix{
		r:    []run{{}},
		cond: sync.NewCond(new(sync.Mutex)),
	}
	return ss
}

// Create a RunScheduler that will communicate with the global scheduler.
func (p *Prefix) GetRunScheduler() RunScheduler {
	return newRunPrefix(p)
}

// Add the provided prefix to the list of unexplored prefixes.
func (p *Prefix) addRun(r run) {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	p.r = append(p.r, r)

	if len(p.r) == 1 {
		p.cond.Broadcast()
	}
}

// End the run
//
// Decrement the number of ongoing runs
func (p *Prefix) endRun() {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	p.ongoing--
	// Signal on the cond that the ongoing variable has changed
	p.cond.Broadcast()
}

// Get the prefix for the next run
//
// Will block until some prefixes are available.
// If no prefixes are available and there are no ongoing runs it will return nil
func (p *Prefix) getRun() run {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	// If there are no available events wait until there are.
	// If at the same time all runSchedulers are waiting for a new event then there will never be a new available event since there are no runSchedulers that can add a new event
	// All possible runs have therefore been explored and we return nil
	// Use the cond to wait for a change in p.ongoing or p.r
	for len(p.r) == 0 && p.ongoing > 0 {
		p.cond.Wait()
	}
	if len(p.r) == 0 {
		return nil
	}

	// Pop the latest prefix
	r := p.r[len(p.r)-1]
	p.r = p.r[:len(p.r)-1]

	p.ongoing++
	return r
}

// Reset the global state of the GlobalScheduler.
// Prepare the scheduler for the next simulation.
func (p *Prefix) Reset() {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	p.r = []run{{}}
	p.ongoing = 0
}

// Manages the exploration of the state space in a single goroutine.
// Events can safely be added from multiple goroutines.
// Events will only be retrieved from a single goroutine during the simulation.
// Communicates with the GlobalScheduler to ensure that the state exploration remains consistent.
type runPrefix struct {
	sync.Mutex

	p *Prefix

	currentIndex int
	currentRun   run

	pendingEvents []event.Event
}

// Create a new runPrefixScheduler
func newRunPrefix(p *Prefix) *runPrefix {
	return &runPrefix{
		p: p,

		currentIndex:  0,
		currentRun:    make(run, 0),
		pendingEvents: make([]event.Event, 0),
	}
}

// Get the next event in the run. Will return RunEndedError if there are no more events in the run.
func (rp *runPrefix) GetEvent() (event.Event, error) {
	rp.Lock()
	defer rp.Unlock()

	if len(rp.pendingEvents) == 0 {
		return nil, RunEndedError
	}

	var evt event.Event
	if rp.currentIndex < len(rp.currentRun) {
		// Follow the current run until it has no more events
		evtId := rp.currentRun[rp.currentIndex]
		// Remove events from the pending events queue as they are selected
		evt = rp.popEvent(evtId)
		if evt == nil {
			return nil, errors.New("Scheduler: Scheduled an event that was pending")
		}
	} else {
		// Pop the last element from the pending events
		evt = rp.pendingEvents[len(rp.pendingEvents)-1]
		rp.pendingEvents = rp.pendingEvents[:len(rp.pendingEvents)-1]

		// For all pending events we create a new run and add it to the pending runs queue
		for _, pendingEvt := range rp.pendingEvents {
			// Add these runs to the pending run slice
			newRun := make(run, len(rp.currentRun))
			copy(newRun, rp.currentRun)
			newRun = append(newRun, pendingEvt.Id())
			rp.p.addRun(newRun)
		}
		rp.currentRun = append(rp.currentRun, evt.Id())
	}
	rp.currentIndex++
	return evt, nil
}

// Get the event from the pending events
// Return nil if it is not found in the pending events
func (rp *runPrefix) popEvent(evtId event.EventId) event.Event {
	for i, pendingEvt := range rp.pendingEvents {
		if evtId == pendingEvt.Id() {
			rp.pendingEvents = append(rp.pendingEvents[:i], rp.pendingEvents[i+1:]...)
			return pendingEvt
		}
	}
	return nil
}

// Implements the event adder interface.
//
// It must be safe to add events from different goroutines.
// StartRun, EndRun and GetEvent will always be called from the same goroutine,
// but not from the same goroutine as AddEvent.
func (rp *runPrefix) AddEvent(evt event.Event) {
	rp.Lock()
	defer rp.Unlock()
	rp.pendingEvents = append(rp.pendingEvents, evt)
}

// Prepare for starting a new run.
//
// Returns a NoRunsError if all possible runs have been completed.
// May block until new runs are available.
// StartRun, EndRun and GetEvent will always be called from the same goroutine,
// but not from the same goroutine as AddEvent.
func (rp *runPrefix) StartRun() error {
	rp.Lock()
	defer rp.Unlock()

	rp.currentIndex = 0
	r := rp.p.getRun()
	if r == nil {
		return NoRunsError
	}
	rp.currentRun = r
	return nil
}

// Finish the current run and prepare for the next one.
//
// Will always be called after a run has been completely executed,
// even if an error occurred during execution of the run.
// StartRun, EndRun and GetEvent will always be called from the same goroutine,
// but not from the same goroutine as AddEvent.
func (rp *runPrefix) EndRun() {
	rp.p.endRun()
}
