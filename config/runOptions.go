package config

import (
	"gomc/failureManager"
	"io"
)

// Configures the Failure Manager that will be used during the simulation

// The Failure Manager controls what crash abstractions will be represented by the simulation
// and which nodes will crash at some point during the simulation.
// The Failure Manager also works as a Failure Detector that informs nodes about status changes.
// Different Failure Managers can represent different types of Failure Detectors.
// Default value is no node crashes.
type FailureManagerOption[T any] struct {
	Fm failureManager.FailureManger[T]
}

func (fmo FailureManagerOption[T]) RunOpt() {}

// Configures io.writers that the discovered state will be exported to

// Can be applied multiple times to add multiple io.writers.
// Default value is no writers.
type ExportOption struct {
	W io.Writer
}

func (eo ExportOption) RunOpt() {}

// Configures a function to shut down a node after the execution of a run.

// The function should clean up any operations to avoid memory leaks across runs.
// Default value is an empty function.
type StopOption[T any] struct {
	Stop func(*T)
}

func (so StopOption[T]) RunOpt() {}

func (so StopOption[T]) RunnerOpt() {}
