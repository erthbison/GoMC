package scheduler

import (
	"gomc/event"
	"math/rand"
	"sync"
)

// A scheduler that randomly picks the next event from the available events.
//
// It is useful for testing a random selection of the state space when the state space is to large to perform an exhaustive search
// It provides no guarantee that all errors have been found, but since it is random it generally contains a larger spread in the states that are checked compared to the exhaustive search.
type Random struct {
	rand *rand.Rand
}

// Create a new Random scheduler
//
// Initialized with a seed which is used to initialize run-specific schedulers
func NewRandom(seed int64) *Random {
	// The provided seed is used to generate seeds for the runScheduler
	return &Random{
		rand: rand.New(rand.NewSource(seed)),
	}
}

// Create a RunScheduler that will communicate with the global scheduler
func (r *Random) GetRunScheduler() RunScheduler {
	return newRandomRun(rand.Int63())
}

// Reset the global state of the GlobalScheduler.
// Prepare the scheduler for the next simulation.
func (r *Random) Reset() {

}

// Manages the exploration of the state space in a single goroutine.
// Events can safely be added from multiple goroutines.
// Events will only be retrieved from a single goroutine during the simulation.
// Communicates with the GlobalScheduler to ensure that the state exploration remains consistent.
type randomRun struct {
	sync.Mutex

	// a slice of all events that can be chosen
	pendingEvents []event.Event

	rand *rand.Rand
}

// Create a new randomRun scheduler
func newRandomRun(seed int64) *randomRun {
	return &randomRun{
		pendingEvents: make([]event.Event, 0),

		rand: rand.New(rand.NewSource(seed)),
	}
}

// Get the next event in the run.
//
// Randomly selects the next event.
//
// Will return RunEndedError if there are no more events in the run.
// The event returned must be an event that has been added during the current run.
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

// Implements the event adder interface.
//
// It must be safe to add events from different goroutines.
// StartRun, EndRun and GetEvent will always be called from the same goroutine,
// but not from the same goroutine as AddEvent.
func (rs *randomRun) AddEvent(evt event.Event) {
	rs.Lock()
	defer rs.Unlock()
	rs.pendingEvents = append(rs.pendingEvents, evt)
}

// Prepare for starting a new run.
//
// Returns a NoRunsError if all possible runs have been completed.
// May block until new runs are available.
// StartRun, EndRun and GetEvent will always be called from the same goroutine,
// but not from the same goroutine as AddEvent.
func (rs *randomRun) StartRun() error {
	rs.Lock()
	defer rs.Unlock()
	rs.pendingEvents = make([]event.Event, 0)
	return nil
}

// Finish the current run and prepare for the next one.
//
// Will always be called after a run has been completely executed,
// even if an error occurred during execution of the run.
// StartRun, EndRun and GetEvent will always be called from the same goroutine,
// but not from the same goroutine as AddEvent.
func (rs *randomRun) EndRun() {
}
