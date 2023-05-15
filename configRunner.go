package gomc

import (
	"gomc/config"
	"gomc/runner"
)

// Configure and start a Runner
//
// The runner is started and then returned.
// Commands can be given to the nodes through the runner.
// See Runner for an overview of the available commands.
//
// Initializes the Runner with the necessary parameters.
// See the RunnerOption for a full overview of possible options.
// Default values will be used if no value is provided.
func PrepareRunner[T, S any](initNodes InitNodeOption[T], getState GetStateOption[T, S], opts ...RunnerOption) *runner.Runner[T, S] {
	var (
		// Method that will be used to
		stop = func(*T) {}

		eventChanBuffer  = 100
		recordChanBuffer = 100
	)

	for _, opt := range opts {
		switch t := opt.(type) {
		case config.StopOption[T]:
			stop = t.Stop
		case config.EventChanBufferOption:
			eventChanBuffer = t.Size
		case config.RecordChanBufferOption:
			recordChanBuffer = t.Size
		}
	}

	r := runner.NewRunner[T, S](
		recordChanBuffer,
	)

	r.Start(
		initNodes.f,
		getState.getState,
		stop,
		eventChanBuffer,
	)
	return r
}

// Specify the function used to collect local state from the nodes
type GetStateOption[T, S any] struct {
	getState func(*T) S
}

// Specify the function used to collect local state from the nodes
func WithStateFunction[T, S any](f func(*T) S) GetStateOption[T, S] {
	return GetStateOption[T, S]{getState: f}
}

// Optional parameters to configure the runner
type RunnerOption interface {
	RunnerOpt()
}

// Configure the buffer size of the records channel
//
// Default value is 100
func RecordChanSize(size int) RunnerOption {
	return config.RecordChanBufferOption{Size: size}
}

// Configure the buffer size of the event channel used to store pending events
//
// Default value is 100
func EventChanBufferSize(size int) RunnerOption {
	return config.EventChanBufferOption{Size: size}
}

// Configures a function used to stop the nodes after a run.
//
// The function should clean up all operations of the nodes to avoid memory leaks across runs.
//
// Default value is empty function.
func WithStopFunctionRunner[T any](stop func(*T)) RunnerOption {
	return config.StopOption[T]{Stop: stop}
}
