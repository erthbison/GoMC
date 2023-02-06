package tester

import (
	"errors"
	"experimentation/tree"
)

type Scheduler[T any] interface {
	GetEvent() (Event[T], error) // Get the next event in the run
	AddEvent(Event[T])           // Add an event to the list of possible events
	EndRun()                     // Finish the current run and prepare for the next one
}

var (
	RunEndedError = errors.New("scheduler: The run has ended. Reset the state.")
	NoEventError  = errors.New("scheduler: No available next event")
)

type BasicScheduler[T any] struct {
	EventRoot    tree.Tree[Event[T]]
	currentEvent *tree.Tree[Event[T]]

	pendingEvents []Event[T]
}

func NewBasicScheduler[T any]() *BasicScheduler[T] {
	eventTree := tree.New[Event[T]](StartEvent[T]{}, EventsEquals[T])
	return &BasicScheduler[T]{
		EventRoot:     eventTree,
		currentEvent:  &eventTree,
		pendingEvents: make([]Event[T], 0),
	}
}

func (bs *BasicScheduler[T]) GetEvent() (Event[T], error) {
	if len(bs.pendingEvents) == 0 {
		return nil, RunEndedError
	}

	// Create new branching paths for all pending events from the current event
	for _, event := range bs.pendingEvents {
		if !bs.currentEvent.HasChild(event) {
			bs.currentEvent.AddChild(event)
		}
	}

	for _, child := range bs.currentEvent.Children {
		// iteratively check if each child can be the next event
		// a child can be the next event if it has some descendent leaf node that is not an "End" event
		if child.SearchLeafNodes(func(e Event[T]) bool { _, ok := e.(EndEvent[T]); return !ok }) {
			bs.removeEvent(child.Payload)
			bs.currentEvent = child
			return child.Payload, nil
		}
	}
	return nil, NoEventError
}

func (bs *BasicScheduler[T]) removeEvent(event Event[T]) {
	// Remove the message from the message queue
	for i, evt := range bs.pendingEvents {
		if EventsEquals(event, evt) {
			bs.pendingEvents = append(bs.pendingEvents[0:i], bs.pendingEvents[i+1:]...)
			break
		}
	}
}

func (bs *BasicScheduler[T]) AddEvent(event Event[T]) {
	bs.pendingEvents = append(bs.pendingEvents, event)
}

func (bs *BasicScheduler[T]) EndRun() {
	// Add an "End" event to the end of the chain
	// Then change the current event to the root of the event tree
	bs.currentEvent.AddChild(EndEvent[T]{})
	bs.currentEvent = &bs.EventRoot
}
