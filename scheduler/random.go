package scheduler

import (
	"gomc/event"
	"math/rand"
	"sync"
)

// A scheduler that randomly picks the next event from the available events.
// It is useful for testing a random selection of the state space when the state space is to large to perform an exhaustive search
// It provides no guarantee that all errors have been found, but since it is random it generally contains a larger spread in the states that are checked compared to the exhaustive search.
type Random struct {
	rand *rand.Rand
}

func NewRandom(seed int64) *Random {
	// The provided seed is used to generate seeds for the runScheduler
	return &Random{
		rand: rand.New(rand.NewSource(seed)),
	}
}

func (r *Random) GetRunScheduler() RunScheduler {
	return newRandomRun(rand.Int63())
}

func (r *Random) Reset() {

}

type randomRun struct {
	sync.Mutex

	// a slice of all events that can be chosen
	pendingEvents []event.Event

	rand *rand.Rand
}

func newRandomRun(seed int64) *randomRun {
	return &randomRun{
		pendingEvents: make([]event.Event, 0),

		rand: rand.New(rand.NewSource(seed)),
	}
}

func (rs *randomRun) GetEvent() (event.Event, error) {
	rs.Lock()
	defer rs.Unlock()

	if len(rs.pendingEvents) == 0 {
		return nil, RunEndedError
	}

	index := rs.rand.Intn(len(rs.pendingEvents))
	evt := rs.pendingEvents[index]

	// Remove the element from the slice. This changes the order of the element in the slice, but is more efficient as we do not need to move all elements after the removed element.
	// Since we are drawing randomly the ordering does not matter
	rs.pendingEvents[index] = rs.pendingEvents[len(rs.pendingEvents)-1]
	rs.pendingEvents = rs.pendingEvents[:len(rs.pendingEvents)-1]

	return evt, nil
}

func (rs *randomRun) AddEvent(evt event.Event) {
	rs.Lock()
	defer rs.Unlock()
	rs.pendingEvents = append(rs.pendingEvents, evt)
}

func (rs *randomRun) StartRun() error {
	rs.Lock()
	defer rs.Unlock()
	rs.pendingEvents = make([]event.Event, 0)
	return nil
}

func (rs *randomRun) EndRun() {
}
