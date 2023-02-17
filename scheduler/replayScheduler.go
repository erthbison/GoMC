package scheduler

import (
	"errors"
	"gomc/event"
)

type replayScheduler struct {
	// A slice of the run to be replayed with event ids in order
	run []string
	// The index of the current event
	index int

	pendingEvents []event.Event
	failed        map[int]bool
}

func NewReplayScheduler(run []string) *replayScheduler {
	return &replayScheduler{
		index: 0,
		run:   run,

		pendingEvents: make([]event.Event, 0),
		failed:        make(map[int]bool),
	}
}

// Get the next event in the run. Will return RunEndedError if there are no more events in the run. Will return NoEventError if there are no more available events in any run.
func (rs *replayScheduler) GetEvent() (event.Event, error) {
	if rs.index >= len(rs.run) {
		return nil, NoEventError
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
	if !rs.failed[evt.Target()] {
		rs.pendingEvents = append(rs.pendingEvents, evt)
	}
}

// Finish the current run and prepare for the next one
func (rs *replayScheduler) EndRun() {
	rs.index = 0
	rs.failed = make(map[int]bool)
}

// Signal to the scheduler that a node has crashed and that events targeting the node should not be scheduled
func (rs *replayScheduler) NodeCrash(id int) {
	// Remove all events that target the node from pending events
	i := 0
	for _, evt := range rs.pendingEvents {
		if evt.Target() != id {
			rs.pendingEvents[i] = evt
			i++
		}
	}
	rs.pendingEvents = rs.pendingEvents[:i]

	rs.failed[id] = true
}
