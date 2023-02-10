package gomc_test

import (
	"gomc"
	"testing"
)

func TestSimulatorNoEvents(t *testing.T) {
	// Test
	sch := NewMockScheduler()
	sm := NewMockStateManager()
	simulator := gomc.NewSimulator[node, state](sch, sm)
	err := simulator.Simulate(func() map[int]*node {
		return map[int]*node{0: {}}
	})
	if err == nil {
		t.Errorf("Expected to receive an error when not providing any functions to simulate")
	}
}
