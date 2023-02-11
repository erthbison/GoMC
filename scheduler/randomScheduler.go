package scheduler

import (
	"gomc/event"
	"math/rand"
)

// A scheduler that randomly picks the next event from the available events.
// It is useful for testing a random selection of the state space when the state space is to large to perform an exhaustive search
// It provides no guarantee that all errors have been found, but since it is random it generally contains a larger spread in the states that are checked compared to the exhaustive search.
type RandomScheduler[T any] struct {
	// a slice of all events that can be chosen
	pendingEvents []event.Event[T]
	numRuns       int
	maxRuns       int

	failed map[int]bool
}

func NewRandomScheduler[T any](maxRuns int) *RandomScheduler[T] {
	return &RandomScheduler[T]{
		pendingEvents: make([]event.Event[T], 0),
		numRuns:       0,
		maxRuns:       maxRuns,

		failed: make(map[int]bool),
	}
}

func (rs *RandomScheduler[T]) GetEvent() (event.Event[T], error) {
	if len(rs.pendingEvents) == 0 {
		return nil, RunEndedError
	}
	if rs.numRuns >= rs.maxRuns {
		return nil, NoEventError
	}

	index := rand.Intn(len(rs.pendingEvents))
	evt := rs.pendingEvents[index]

	// Remove the element from the slice. This changes the order of the element in the slice, but is more efficient as we do not need to move all elements after the removed element.
	// Since we are drawing randomly the ordering does not matter
	rs.pendingEvents[index] = rs.pendingEvents[len(rs.pendingEvents)-1]
	rs.pendingEvents = rs.pendingEvents[:len(rs.pendingEvents)-1]

	return evt, nil
}

func (rs *RandomScheduler[T]) AddEvent(evt event.Event[T]) {
	if rs.failed[evt.Target()] {
		return
	}
	rs.pendingEvents = append(rs.pendingEvents, evt)
}

func (rs *RandomScheduler[T]) EndRun() {
	rs.numRuns++
	// The pendingEvents slice is supposed to be empty when the run ends, but just in case it is not(or the run is manually reset), create a new, empty slice.
	rs.pendingEvents = make([]event.Event[T], 0)

	// Reset the map of failed nodes
	rs.failed = make(map[int]bool)
}

func (rs *RandomScheduler[T]) NodeCrash(id int) {
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
