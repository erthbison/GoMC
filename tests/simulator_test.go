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
	err := simulator.Simulate(
		func() map[int]*node {
			return map[int]*node{0: {}}
		},
		map[int][]func(*node) error{
			0: {},
		},
		[]int{},
	)
	if err == nil {
		t.Errorf("Expected to receive an error when not providing any functions to simulate")
	}

	err = simulator.Simulate(
		func() map[int]*node {
			return map[int]*node{0: {}}
		},
		map[int][]func(*node) error{},
		[]int{},
	)
	if err == nil {
		t.Errorf("Expected to receive an error when not providing any functions to simulate")
	}
}
