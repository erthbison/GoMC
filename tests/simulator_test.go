package gomc_test

import (
	"gomc"
	"testing"
)

// var events = []gomc.Event[node]{}

func TestSimulatorNoEvents(t *testing.T) {
	sch := NewMockScheduler()
	sm := NewMockStateManager()
	simulator := gomc.NewSimulator[node, state](sch, sm)
	simulator.Simulate(func() map[int]*node {
		return map[int]*node{0: {}}
	})
}
