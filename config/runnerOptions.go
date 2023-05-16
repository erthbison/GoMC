package config

// Configures how many records can be buffered

// Default value is 100
type RecordChanBufferOption struct {
	Size int
}

func (opt RecordChanBufferOption) RunnerOpt() {}

// Configures how many pending event can be buffered by each node

// Default value is 100
type EventChanBufferOption struct {
	Size int
}

func (opt EventChanBufferOption) RunnerOpt() {}
