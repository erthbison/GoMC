package runner

import "reflect"

type command interface {
	cmd()
}

// Pause the execution of the node with the provided Id
type pauseCmd struct {
	Id int
}

func (pc pauseCmd) cmd() {}

// Resume the execution of the node with the provided Id
type resumeCmd struct {
	Id int
}

func (pc resumeCmd) cmd() {}

// Crash the node with the provided Id
type crashCmd struct {
	Id int
}

func (pc crashCmd) cmd() {}

// Send a request to the node with teh provided Id
type requestCmd struct {
	Id     int
	Method string
	Params []reflect.Value
}

func (pc requestCmd) cmd() {}

// Stop the running of the algorithm
type stopCmd struct{}

func (pc stopCmd) cmd() {}
