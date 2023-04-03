package runner

import "reflect"

type pause struct {
	id int
}

type resume struct {
	id int
}

type crash struct {
	id int
}

type request struct {
	id     int
	method string
	params []reflect.Value
}

type stop struct{}
