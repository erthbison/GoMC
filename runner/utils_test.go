package runner

import (
	"gomc/event"
)

// Create some dummy types and states for use when testing
type MockNode struct {
	Id int

	crashed bool
	val     int
}

func (n *MockNode) UpdateVal(val int) {
	n.val = val
}

func GetState(n *MockNode) State {
	return State{
		val: n.val,
	}
}

type State struct {
	val int
}

type MockEvent struct {
	id     event.EventId
	target int
	val    int
}

func (me MockEvent) Id() event.EventId {
	return me.id
}

func (me MockEvent) Execute(n any, chn chan error) {
	if me.val == -1 {
		panic("Node panicked during testing")
	}
	tmp := n.(*MockNode)
	if !tmp.crashed {
		tmp.val = me.val
	}
	chn <- nil
}

func (me MockEvent) Target() int {
	return me.target
}
