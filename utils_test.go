package gomc

import (
	"strconv"
)

// Create some dummy types and states for use when testing
type MockNode struct {
	Id int
}

func (n *MockNode) Foo(from, to int, msg []byte) {}
func (n *MockNode) Bar(from, to int, msg []byte) {}

type MockEvent struct {
	id       int
	target   int
	executed bool
}

func (me MockEvent) Id() string {
	return strconv.Itoa(me.id)
}

func (me MockEvent) Execute(_ any, chn chan error) {
	chn <- nil
}

func (me MockEvent) Target() int {
	return me.target
}
