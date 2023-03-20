package scheduler

import (
	"errors"
	"gomc/event"
	"sync"
)

type replayScheduler struct {
	sync.Mutex
	// A slice of the run to be replayed with event ids in order
	run []string
	// The index of the current event
	index int

	done          bool
	pendingEvents []event.Event
}

func NewReplayScheduler(run []string) *replayScheduler {
	return &replayScheduler{
		index: 0,
		run:   run,

		pendingEvents: make([]event.Event, 0),

		done: false,
	}
}

// Get the next event in the run. Will return RunEndedError if there are no more events in the run. Will return NoEventError if there are no more available events in any run.
func (rs *replayScheduler) GetEvent() (event.Event, error) {
	rs.Lock()
	defer rs.Unlock()
	if rs.done {
		return nil, NoEventError
	}
	if rs.index >= len(rs.run) {
		rs.done = true
		return nil, RunEndedError
	}
	evtId := rs.run[rs.index]
	evt := rs.popEvent(evtId)
	if evt == nil {
		return nil, errors.New("Unable to find next event")
	}
	rs.index++
	return evt, nil
}

func (rs *replayScheduler) popEvent(id string) event.Event {
	for i, evt := range rs.pendingEvents {
		if evt.Id() == id {
			// Remove the event from the pending events and return the event
			rs.pendingEvents = append(rs.pendingEvents[:i], rs.pendingEvents[i+1:]...)
			return evt
		}
	}
	return nil
}

// Add an event to the list of possible events
func (rs *replayScheduler) AddEvent(evt event.Event) {
	rs.Lock()
	defer rs.Unlock()
	rs.pendingEvents = append(rs.pendingEvents, evt)
}

// Finish the current run and prepare for the next one
func (rs *replayScheduler) EndRun() {
	rs.Lock()
	defer rs.Unlock()
	rs.index = 0
	// rs.done = false
}
