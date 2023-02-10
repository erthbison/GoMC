package scheduler

import (
	"gomc/event"
	"math/rand"
)

type RandomScheduler[T any] struct {
	pendingEvents []event.Event[T]
	numRuns       int
	maxRuns       int
}

func NewRandomScheduler[T any](maxRuns int) *RandomScheduler[T] {
	return &RandomScheduler[T]{
		pendingEvents: make([]event.Event[T], 0),
		numRuns:       0,
		maxRuns:       maxRuns,
	}
}

func (rs *RandomScheduler[T]) GetEvent() (event.Event[T], error) {
	if len(rs.pendingEvents) == 0 {
		return nil, RunEndedError
	}
	if rs.numRuns > rs.maxRuns {
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
	rs.pendingEvents = append(rs.pendingEvents, evt)
}

func (rs *RandomScheduler[T]) EndRun() {
	rs.numRuns++
	// The pendingEvents slice is supposed to be empty when the run ends, but just in case it is not(or the run is manually reset), create a new, empty slice.
	rs.pendingEvents = make([]event.Event[T], 0)
}
