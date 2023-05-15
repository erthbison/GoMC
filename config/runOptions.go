package config

import (
	"gomc/failureManager"
	"io"
)

type FailureManagerOption[T any] struct {
	Fm failureManager.FailureManger[T]
}

func (fmo FailureManagerOption[T]) RunOpt() {}

type ExportOption struct {
	W io.Writer
}

func (eo ExportOption) RunOpt() {}

// Configure a function to shut down a node after the execution of a run.
// Default value is an empty function.
type StopOption[T any] struct {
	Stop func(*T)
}

func (so StopOption[T]) RunOpt() {}

func (so StopOption[T]) RunnerOpt() {}
