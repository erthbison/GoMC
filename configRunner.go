package gomc

import (
	"gomc/config"
	"gomc/runner"
)

func PrepareRunner[T, S any](initNodes InitNodeOption[T], getState GetStateOption[T, S], opts ...RunnerOption) *runner.Runner[T, S] {
	var (
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

type GetStateOption[T, S any] struct {
	getState func(*T) S
}

func WithStateFunction[T, S any](f func(*T) S) GetStateOption[T, S] {
	return GetStateOption[T, S]{getState: f}
}

type RunnerOption interface {
	RunnerOpt()
}

func RecordChanSize(size int) RunnerOption {
	return config.RecordChanBufferOption{Size: size}
}

func EventChanBufferSize(size int) RunnerOption {
	return config.EventChanBufferOption{Size: size}
}
