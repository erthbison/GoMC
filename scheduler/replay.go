package scheduler

import (
	"errors"
	"gomc/event"
	"sync"
)

type Replay struct {
	run  []string
	done bool
}

func NewReplay(run []string) *Replay {
	return &Replay{
		run: run,
	}
}

func (r *Replay) GetRunScheduler() RunScheduler {
	if r.done {
		return newRunReplay(nil)
	}
	r.done = true
	return newRunReplay(r.run)
}

type runReplay struct {
	sync.Mutex
	// A slice of the run to be replayed with event ids in order
	run []string
	// The index of the current event
	index int

	pendingEvents []event.Event
}

func newRunReplay(run []string) *runReplay {
	return &runReplay{
		index: 0,
		run:   run,

		pendingEvents: make([]event.Event, 0),
	}
}

// Get the next event in the run. Will return RunEndedError if there are no more events in the run. Will return NoEventError if there are no more available events in any run.
func (rr *runReplay) GetEvent() (event.Event, error) {
	rr.Lock()
	defer rr.Unlock()

	if rr.index >= len(rr.run) {
		rr.run = nil
		return nil, RunEndedError
	}
	evtId := rr.run[rr.index]
	evt := rr.popEvent(evtId)
	if evt == nil {
		return nil, errors.New("Unable to find next event")
	}
	rr.index++
	return evt, nil
}

func (rr *runReplay) popEvent(id string) event.Event {
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
func (rr *runReplay) AddEvent(evt event.Event) {
	rr.Lock()
	defer rr.Unlock()
	rr.pendingEvents = append(rr.pendingEvents, evt)
}

func (rr *runReplay) StartRun() error {
	rr.Lock()
	defer rr.Unlock()
	if rr.run == nil {
		return NoRunsError
	}
	return nil
}

// Finish the current run and prepare for the next one
func (rr *runReplay) EndRun() {
	rr.Lock()
	defer rr.Unlock()
	rr.index = 0
}
