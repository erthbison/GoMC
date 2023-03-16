package scheduler

import (
	"gomc/event"
	"math/rand"
)

// A scheduler that randomly picks the next event from the available events.
// It is useful for testing a random selection of the state space when the state space is to large to perform an exhaustive search
// It provides no guarantee that all errors have been found, but since it is random it generally contains a larger spread in the states that are checked compared to the exhaustive search.
type RandomScheduler struct {
	// a slice of all events that can be chosen
	pendingEvents []event.Event
	numRuns       uint
	maxRuns       uint

	rand *rand.Rand
}

func NewRandomScheduler(maxRuns uint, seed int64) *RandomScheduler {
	rand := rand.New(rand.NewSource(1))
	return &RandomScheduler{
		pendingEvents: make([]event.Event, 0),
		numRuns:       0,
		maxRuns:       maxRuns,

		rand: rand,
	}
}

func (rs *RandomScheduler) GetEvent() (event.Event, error) {
	if len(rs.pendingEvents) == 0 {
		return nil, RunEndedError
	}
	if rs.numRuns >= rs.maxRuns {
		return nil, NoEventError
	}

	index := rs.rand.Intn(len(rs.pendingEvents))
	evt := rs.pendingEvents[index]

	// Remove the element from the slice. This changes the order of the element in the slice, but is more efficient as we do not need to move all elements after the removed element.
	// Since we are drawing randomly the ordering does not matter
	rs.pendingEvents[index] = rs.pendingEvents[len(rs.pendingEvents)-1]
	rs.pendingEvents = rs.pendingEvents[:len(rs.pendingEvents)-1]

	return evt, nil
}

func (rs *RandomScheduler) AddEvent(evt event.Event) {
	rs.pendingEvents = append(rs.pendingEvents, evt)
}

func (rs *RandomScheduler) EndRun() {
	rs.numRuns++
	// The pendingEvents slice is supposed to be empty when the run ends, but just in case it is not(or the run is manually reset), create a new, empty slice.
	rs.pendingEvents = make([]event.Event, 0)
}
