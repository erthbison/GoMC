package scheduler

import (
	"errors"
	"gomc/event"
	"sync"
)

type run []event.EventId

// Explores the state space by maintaining a stack of unexplored prefixes.
// When a new run is started it follows the prefix and begins exploring from there, adding new prefixes it discovers as it executes events.
type Prefix struct {
	// unexplored prefixes
	r []run

	// Used to wait for a change in p.ongoing or p.r. The condition is len(p.r) == 0 and p.ongoing > 0
	cond *sync.Cond

	// Number of runScheduler currently scheduling a run. I.e. runScheduler no waiting for a new run
	ongoing int
}

func NewPrefix() *Prefix {
	ss := &Prefix{
		r:    []run{{}},
		cond: sync.NewCond(new(sync.Mutex)),
	}
	return ss
}

func (p *Prefix) GetRunScheduler() RunScheduler {
	return newRunQueue(p)
}

func (p *Prefix) addRun(r run) {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	p.r = append(p.r, r)

	if len(p.r) == 1 {
		p.cond.Broadcast()
	}
}

func (p *Prefix) endRun() {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	p.ongoing--
	p.cond.Broadcast()
}

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

type runPrefix struct {
	sync.Mutex

	p *Prefix

	currentIndex int
	currentRun   run

	pendingEvents []event.Event

	pendingRuns []run
}

func newRunQueue(p *Prefix) *runPrefix {
	return &runPrefix{
		p: p,

		currentIndex:  0,
		currentRun:    make(run, 0),
		pendingEvents: make([]event.Event, 0),
		pendingRuns:   make([]run, 0),
	}
}

// Get the next event in the run. Will return RunEndedError if there are no more events in the run. Will return NoEventError if there are no more available events in any run.
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

func (rp *runPrefix) popEvent(evtId event.EventId) event.Event {
	// Remove the message from the message queue
	for i, pendingEvt := range rp.pendingEvents {
		if evtId == pendingEvt.Id() {
			rp.pendingEvents = append(rp.pendingEvents[:i], rp.pendingEvents[i+1:]...)
			return pendingEvt
		}
	}
	return nil
}

// Add an event to the list of possible events
func (rp *runPrefix) AddEvent(evt event.Event) {
	rp.Lock()
	defer rp.Unlock()
	rp.pendingEvents = append(rp.pendingEvents, evt)
}

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

// Finish the current run and prepare for the next one
func (rp *runPrefix) EndRun() {
	rp.p.endRun()
}
