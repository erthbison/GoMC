package runner

import "reflect"

type Pause struct {
	Id int
}

type Resume struct {
	Id int
}

type Crash struct {
	Id int
}

type Request struct {
	Id     int
	Method string
	Params []reflect.Value
}

type Stop struct{}
