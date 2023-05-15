package config

type RecordChanBufferOption struct {
	Size int
}

func (opt RecordChanBufferOption) RunnerOpt() {}

type EventChanBufferOption struct {
	Size int
}

func (opt EventChanBufferOption) RunnerOpt() {}
