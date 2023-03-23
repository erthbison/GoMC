package scheduler

import (
	"errors"
	"gomc/event"
)

type GlobalScheduler interface {
	// Used to manage the exploration of the state space.
	// The global scheduler manages the total state across several runs.
	// Communicates with several run schedulers in separate goroutines to ensure that the exploration remains consistent

	// Create a RunScheduler that will communicate with the global scheduler
	GetRunScheduler() RunScheduler
}

type RunScheduler interface {
	// Manages the exploration of the state space in a single goroutine.
	// Events can safely be added from multiple goroutines.
	// Events will only be retrieved from a single goroutine during the simulation.
	// Communicates with the GlobalScheduler to ensure that the state exploration remains consistent.

	// Get the next event in the run. Will return RunEndedError if there are no more events in the run.
	// The event returned must be an event that has been added during the current run.
	GetEvent() (event.Event, error)
	// Add an event to the list of possible events
	AddEvent(event.Event)

	// Prepare for starting a new run. Returns a NoRunsError if all possible runs have been completed. May block until new runs are available.
	StartRun() error
	// Finish the current run and prepare for the next one
	EndRun()
}

var (
	RunEndedError = errors.New("scheduler: The run has ended. Reset the state.")
	NoRunsError   = errors.New("scheduler: No available new runs to be started")
)
