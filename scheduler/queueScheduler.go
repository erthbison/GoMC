package scheduler

import (
	"errors"
	"gomc/event"
)

type QueueScheduler struct {
	currentIndex int
	currentRun   []string

	pendingRuns [][]string

	pendingEvents []event.Event
	crashedNodes  map[int]bool
}

func NewQueueScheduler() *QueueScheduler {
	return &QueueScheduler{
		currentIndex:  0,
		currentRun:    make([]string, 0),
		pendingRuns:   make([][]string, 0),
		pendingEvents: make([]event.Event, 0),
		crashedNodes:  make(map[int]bool),
	}
}

// Get the next event in the run. Will return RunEndedError if there are no more events in the run. Will return NoEventError if there are no more available events in any run.
func (qs *QueueScheduler) GetEvent() (event.Event, error) {
	if qs.currentRun == nil {
		return nil, NoEventError
	}
	if len(qs.pendingEvents) == 0 {
		return nil, RunEndedError
	}

	var evt event.Event
	if qs.currentIndex < len(qs.currentRun) {
		// Follow the current run until it has no more events
		evtId := qs.currentRun[qs.currentIndex]
		// Remove events from the pending events queue as they are selected
		evt = qs.popEvent(evtId)
		if evt == nil {
			return nil, errors.New("Scheduler: Scheduled an event that was pending")
		}
	} else {
		// Pop the last element from the pending events
		evt = qs.pendingEvents[len(qs.pendingEvents)-1]
		qs.pendingEvents = qs.pendingEvents[:len(qs.pendingEvents)-1]

		// For all other events in the pending events queue we create a new run and add it to the pending runs queue
		for _, pendingEvt := range qs.pendingEvents {
			// Add these runs to the pending run slice
			run := make([]string, len(qs.currentRun))
			copy(run, qs.currentRun)
			run = append(run, pendingEvt.Id())
			qs.pendingRuns = append(qs.pendingRuns, run)
		}
		qs.currentRun = append(qs.currentRun, evt.Id())
	}
	qs.currentIndex++
	return evt, nil
}

func (qs *QueueScheduler) popEvent(evtId string) event.Event {
	// Remove the message from the message queue
	for i, pendingEvt := range qs.pendingEvents {
		if evtId == pendingEvt.Id() {
			qs.pendingEvents = append(qs.pendingEvents[:i], qs.pendingEvents[i+1:]...)
			return pendingEvt
		}
	}
	return nil
}

// Add an event to the list of possible events
func (qs *QueueScheduler) AddEvent(evt event.Event) {
	if !qs.crashedNodes[evt.Target()] {
		qs.pendingEvents = append(qs.pendingEvents, evt)
	}
}

// Finish the current run and prepare for the next one
func (qs *QueueScheduler) EndRun() {
	qs.currentIndex = 0
	qs.crashedNodes = make(map[int]bool)
	if len(qs.pendingRuns) > 0 {
		qs.currentRun = qs.pendingRuns[len(qs.pendingRuns)-1]
		qs.pendingRuns = qs.pendingRuns[:len(qs.pendingRuns)-1]
	} else {
		qs.currentRun = nil
	}

}

// Signal to the scheduler that a node has crashed and that events targeting the node should not be scheduled
func (qs *QueueScheduler) NodeCrash(id int) {
	qs.crashedNodes[id] = true

	// remove all pending events with the crashed node as its target
	i := 0
	for _, evt := range qs.pendingEvents {
		if evt.Target() != id {
			qs.pendingEvents[i] = evt
			i++
		}
	}
	qs.pendingEvents = qs.pendingEvents[:i]
}
