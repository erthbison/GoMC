package scheduler

import (
	"errors"
	"gomc/event"
)

type Scheduler[T any] interface {
	// Get the next event in the run. Will return RunEndedError if there are no more events in the run. Will return NoEventError if there are no more available events in any run.
	GetEvent() (event.Event[T], error)
	// Add an event to the list of possible events
	AddEvent(event.Event[T])
	// Finish the current run and prepare for the next one
	EndRun()
}

var (
	RunEndedError = errors.New("scheduler: The run has ended. Reset the state.")
	NoEventError  = errors.New("scheduler: No available next event")
)
