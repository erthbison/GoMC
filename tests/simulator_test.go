package gomc_test

import (
	"gomc"
	"testing"
)

func TestSimulatorNoEvents(t *testing.T) {
	// Test
	sch := NewMockGlobalScheduler()
	sm := NewMockStateManager()
	simulator := gomc.NewSimulator[MockNode, State](sch, false, false, 10000, 1000, 1)
	err := simulator.Simulate(
		sm,
		func(sp gomc.SimulationParameters) map[int]*MockNode {
			return map[int]*MockNode{0: {}}
		},
		[]int{},
		func(*MockNode) {},
	)
	if err == nil {
		t.Errorf("Expected to receive an error when not providing any functions to simulate")
	}

	sm = NewMockStateManager()
	err = simulator.Simulate(
		sm,
		func(sp gomc.SimulationParameters) map[int]*MockNode {
			return map[int]*MockNode{0: {}}
		},
		[]int{},
		func(*MockNode) {},
	)
	if err == nil {
		t.Errorf("Expected to receive an error when not providing any functions to simulate")
	}
}
