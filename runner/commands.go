package runner

import "reflect"

// Pause the execution of the node with the provided Id
type Pause struct {
	Id int
}

// Resume the execution of the node with the provided Id
type Resume struct {
	Id int
}

// Crash the node with the provided Id
type Crash struct {
	Id int
}

// Send a request to the node with teh provided Id
type Request struct {
	Id     int
	Method string
	Params []reflect.Value
}

// Stop the running of the algorithm
type Stop struct{}
