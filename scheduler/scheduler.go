package scheduler

import (
	"errors"
	"gomc/event"
)

type Scheduler[T any] interface {
	GetEvent() (event.Event[T], error) // Get the next event in the run. Will return RunEndedError if there are no more events in the run. Will return NoEventError if there are no more available events in any run.
	AddEvent(event.Event[T])           // Add an event to the list of possible events
	EndRun()                           // Finish the current run and prepare for the next one
}

var (
	RunEndedError = errors.New("scheduler: The run has ended. Reset the state.")
	NoEventError  = errors.New("scheduler: No available next event")
)
