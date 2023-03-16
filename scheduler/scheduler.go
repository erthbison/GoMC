package scheduler

import (
	"errors"
	"gomc/event"
)

type Scheduler interface {
	// Get the next event in the run. Will return RunEndedError if there are no more events in the run. Will return NoEventError if called after returning a RunEnded error if a new run has not been started and if there are no more available events in any run.
	// The event returned must be an event that has been added during the current run.
	GetEvent() (event.Event, error)
	// Add an event to the list of possible events
	AddEvent(event.Event)
	// Finish the current run and prepare for the next one
	EndRun()
}

var (
	RunEndedError = errors.New("scheduler: The run has ended. Reset the state.")
	NoEventError  = errors.New("scheduler: No available next event")
)
