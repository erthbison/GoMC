package scheduler

import (
	"gomc/event"
	"gomc/tree"
)

type BasicScheduler[T any] struct {
	EventRoot    *tree.Tree[event.Event[T]]
	currentEvent *tree.Tree[event.Event[T]]

	// Must be a slice to allow for duplicate entries of messages. If the same message has been sent twice we want it to arrive twice
	pendingEvents []event.Event[T]
}

func NewBasicScheduler[T any]() *BasicScheduler[T] {
	eventTree := tree.New(nil, event.EventsEquals[T])
	return &BasicScheduler[T]{
		EventRoot:     &eventTree,
		currentEvent:  &eventTree,
		pendingEvents: make([]event.Event[T], 0),
	}
}

func (bs *BasicScheduler[T]) GetEvent() (event.Event[T], error) {
	if len(bs.pendingEvents) == 0 {
		return nil, RunEndedError
	}

	// Create new branching paths for all pending events from the current event
	for _, event := range bs.pendingEvents {
		if !bs.currentEvent.HasChild(event) {
			bs.currentEvent.AddChild(event)
		}
	}

	for _, child := range bs.currentEvent.Children() {
		// iteratively check if each child can be the next event
		// a child can be the next event if it has some descendent leaf node that is not an "End" event
		if child.SearchLeafNodes(func(e event.Event[T]) bool { return e != nil }) {
			bs.removeEvent(child.Payload())
			bs.currentEvent = child
			return child.Payload(), nil
		}
	}
	return nil, NoEventError
}

func (bs *BasicScheduler[T]) removeEvent(evt event.Event[T]) {
	// Remove the message from the message queue
	for i, pendingEvt := range bs.pendingEvents {
		if event.EventsEquals(evt, pendingEvt) {
			bs.pendingEvents = append(bs.pendingEvents[0:i], bs.pendingEvents[i+1:]...)
			break
		}
	}
}

func (bs *BasicScheduler[T]) AddEvent(evt event.Event[T]) {
	bs.pendingEvents = append(bs.pendingEvents, evt)
}

func (bs *BasicScheduler[T]) EndRun() {
	// Add an "End" event to the end of the chain
	// Then change the current event to the root of the event tree
	bs.currentEvent.AddChild(nil)
	bs.currentEvent = bs.EventRoot
	// The pendingEvents slice is supposed to be empty when the run ends, but just in case it is not(or the run is manually reset), create a new, empty slice.
	bs.pendingEvents = make([]event.Event[T], 0)
}
