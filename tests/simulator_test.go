package gomc_test

import (
	"gomc"
	"testing"
)

func TestSimulatorNoEvents(t *testing.T) {
	// Test
	sch := NewMockScheduler()
	sm := NewMockStateManager()
	simulator := gomc.NewSimulator[MockNode, State](sch, sm, 10000, 1000)
	err := simulator.Simulate(
		func() map[int]*MockNode {
			return map[int]*MockNode{0: {}}
		},
		[]int{},
	)
	if err == nil {
		t.Errorf("Expected to receive an error when not providing any functions to simulate")
	}

	err = simulator.Simulate(
		func() map[int]*MockNode {
			return map[int]*MockNode{0: {}}
		},
		[]int{},
	)
	if err == nil {
		t.Errorf("Expected to receive an error when not providing any functions to simulate")
	}
}
