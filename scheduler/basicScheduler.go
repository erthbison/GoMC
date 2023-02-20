package scheduler

import (
	"errors"
	"gomc/event"
	"gomc/tree"
)

type BasicScheduler struct {
	// Trees storing the ID of events as the payload
	// Are used to keep track of already executed runs
	EventRoot    *tree.Tree[string]
	currentEvent *tree.Tree[string]

	// Must be a slice to allow for duplicate entries of messages. If the same message has been sent twice we want it to arrive twice
	pendingEvents []event.Event

	failed map[int]bool
}

func NewBasicScheduler() *BasicScheduler {
	eventTree := tree.New("Start", func(a, b string) bool { return a == b })
	return &BasicScheduler{
		EventRoot:     &eventTree,
		currentEvent:  &eventTree,
		pendingEvents: make([]event.Event, 0),

		failed: make(map[int]bool),
	}
}

func (bs *BasicScheduler) GetEvent() (event.Event, error) {
	if len(bs.pendingEvents) == 0 {
		return nil, RunEndedError
	}

	// Create new branching paths for all pending events from the current event
	for _, event := range bs.pendingEvents {
		if !bs.currentEvent.HasChild(event.Id()) {
			bs.currentEvent.AddChild(event.Id())
		}
	}

	for _, child := range bs.currentEvent.Children() {
		// iteratively check if each child can be the next event
		// a child can be the next event if it has some descendent leaf node that is not an "End" event
		if child.SearchLeafNodes(func(n string) bool { return n != "End" }) {
			evt := bs.popEvent(child.Payload())
			if evt == nil {
				return nil, errors.New("Scheduler: Scheduled non-pending event")
			}
			bs.currentEvent = child
			return evt, nil
		}
	}
	return nil, NoEventError
}

func (bs *BasicScheduler) popEvent(evtId string) event.Event {
	// Remove the message from the message queue
	for i, pendingEvt := range bs.pendingEvents {
		if evtId == pendingEvt.Id() {
			bs.pendingEvents = append(bs.pendingEvents[0:i], bs.pendingEvents[i+1:]...)
			return pendingEvt
		}
	}
	return nil
}

func (bs *BasicScheduler) AddEvent(evt event.Event) {
	if bs.failed[evt.Target()] {
		return
	}
	bs.pendingEvents = append(bs.pendingEvents, evt)
}

func (bs *BasicScheduler) EndRun() {
	// Add an "End" event to the end of the chain
	// Then change the current event to the root of the event tree
	bs.currentEvent.AddChild("End")
	bs.currentEvent = bs.EventRoot
	// The pendingEvents slice is supposed to be empty when the run ends, but just in case it is not(or the run is manually reset), create a new, empty slice.
	bs.pendingEvents = make([]event.Event, 0)

	// Reset the map of failed nodes
	bs.failed = make(map[int]bool)
}

func (bs *BasicScheduler) NodeCrash(id int) {
	// Remove all events that target the node from pending events

	i := 0
	for _, evt := range bs.pendingEvents {
		if evt.Target() != id {
			bs.pendingEvents[i] = evt
			i++
		}
	}
	bs.pendingEvents = bs.pendingEvents[:i]

	bs.failed[id] = true
}
