package main

import (
	"bytes"
	"encoding/json"
	"gomc"
	"gomc/eventManager"
	"gomc/scheduler"
	"os"
	"testing"
)

type state struct {
	round    uint
	proposed Value[int]
	decided  Value[int]
}

func TestConsensus(t *testing.T) {
	sch := scheduler.NewQueueScheduler()
	// sch := scheduler.NewRandomScheduler(1000)
	sm := gomc.NewStateManager(
		func(node *HierarchicalConsensus[int]) state {
			var val Value[int]
			select {
			case val = <-node.DecidedSignal:
			default:
			}
			return state{
				proposed: node.proposal,
				decided:  val,
				round:    node.round,
			}
		},
		func(a, b state) bool {
			if a.round != b.round {
				return false
			}
			if a.proposed.val != b.proposed.val {
				return false
			}
			return a.decided.val == b.decided.val
		},
	)
	sender := eventManager.NewSender(sch)
	sim := gomc.NewSimulator[HierarchicalConsensus[int], state](sch, sm, 10000, 1000)
	err := sim.Simulate(
		func() map[int]*HierarchicalConsensus[int] {
			nodeIds := []uint{}
			numNodes := 5
			for i := 1; i <= numNodes; i++ {
				nodeIds = append(nodeIds, uint(i))
			}
			nodes := make(map[int]*HierarchicalConsensus[int])
			for _, id := range nodeIds {
				node := NewHierarchicalConsensus[int](
					id,
					nodeIds,
					sender.SendFunc(int(id)),
				)
				sim.Fm.Subscribe(node.Crash)
				nodes[int(id)] = node
			}
			return nodes
		},
		[]int{1, 3},
		gomc.NewRequest(1, "Propose", Value[int]{2}),
		gomc.NewRequest(2, "Propose", Value[int]{3}),
		gomc.NewRequest(3, "Propose", Value[int]{4}),
		gomc.NewRequest(4, "Propose", Value[int]{5}),
		gomc.NewRequest(5, "Propose", Value[int]{6}),
	)
	if err != nil {
		t.Errorf("Expected no error while simulating. Got %v", err)
	}

	checker := gomc.NewPredicateChecker(
		gomc.PredEventually(
			// C1: termination and C3: Integrity
			func(globalState gomc.GlobalState[state], isTerminal bool, sequence []gomc.GlobalState[state]) bool {
				nodeDecided := make(map[int]int)
				decidedValues := make(map[int]bool)
				proposedValues := make(map[int]bool)
				for _, gs := range sequence {
					for node, state := range gs.LocalStates {
						if state.decided.val != 0 {
							// C3: If some value has previously been decided. And then a new value is decided this breaks integrity
							if nodeDecided[node] != 0 {
								return false
							}
							nodeDecided[node] = state.decided.val
							if globalState.Correct[node] {
								decidedValues[state.decided.val] = true
							}
							proposedValues[state.proposed.val] = true
						}
					}
				}
				// C1: Check that all correct nodes decided on some value
				for node, correct := range globalState.Correct {
					if correct {
						if _, ok := nodeDecided[node]; !ok {
							return false
						}
					}
				}
				// Check that at most one value was decided
				if len(decidedValues) > 1 {
					return false
				}
				// Check that a decided values was at some point proposed
				for decidedVal := range decidedValues {
					if !proposedValues[decidedVal] {
						return false
					}
				}
				return true
			},
		),
	)
	resp := checker.Check(sm.StateRoot)
	if ok, out := resp.Response(); !ok {
		t.Errorf("Expected no errors while checking. Got: %v", out)
	}
	var buffer bytes.Buffer
	json.NewEncoder(&buffer).Encode(resp.Export())
	os.WriteFile("FailedRun.txt", buffer.Bytes(), 0755)
}

func TestConsensusReplay(t *testing.T) {
	in, err := os.ReadFile("FailedRun.txt")
	if err != nil {
		t.Errorf("Error while setting up test: %v", err)
	}
	buffer := bytes.NewBuffer(in)
	var run []string
	json.NewDecoder(buffer).Decode(&run)

	sch := scheduler.NewReplayScheduler(run)
	sm := gomc.NewStateManager(
		func(node *HierarchicalConsensus[int]) state {
			var val Value[int]
			select {
			case val = <-node.DecidedSignal:
			default:
			}
			return state{
				proposed: node.proposal,
				decided:  val,
				round:    node.round,
			}
		},
		func(a, b state) bool {
			if a.round != b.round {
				return false
			}
			if a.proposed.val != b.proposed.val {
				return false
			}
			return a.decided.val == b.decided.val
		},
	)
	sender := eventManager.NewSender(sch)
	sim := gomc.NewSimulator[HierarchicalConsensus[int], state](sch, sm, 10000, 1000)
	err = sim.Simulate(
		func() map[int]*HierarchicalConsensus[int] {
			nodeIds := []uint{}
			numNodes := 5
			for i := 1; i <= numNodes; i++ {
				nodeIds = append(nodeIds, uint(i))
			}
			nodes := make(map[int]*HierarchicalConsensus[int])
			for _, id := range nodeIds {
				node := NewHierarchicalConsensus[int](
					id,
					nodeIds,
					sender.SendFunc(int(id)),
				)
				sim.Fm.Subscribe(node.Crash)
				nodes[int(id)] = node
			}
			return nodes
		},
		[]int{1, 3},
		gomc.NewRequest(1, "Propose", Value[int]{2}),
		gomc.NewRequest(2, "Propose", Value[int]{3}),
		gomc.NewRequest(3, "Propose", Value[int]{4}),
		gomc.NewRequest(4, "Propose", Value[int]{5}),
		gomc.NewRequest(5, "Propose", Value[int]{6}),
	)
	if err != nil {
		t.Errorf("Expected no error while simulating. Got %v", err)
	}

	checker := gomc.NewPredicateChecker(
		gomc.PredEventually(
			// C1: termination and C3: Integrity
			func(globalState gomc.GlobalState[state], isTerminal bool, sequence []gomc.GlobalState[state]) bool {
				nodeDecided := make(map[int]int)
				decidedValues := make(map[int]bool)
				proposedValues := make(map[int]bool)
				for _, gs := range sequence {
					for node, state := range gs.LocalStates {
						if state.decided.val != 0 {
							// C3: If some value has previously been decided. And then a new value is decided this breaks integrity
							if nodeDecided[node] != 0 {
								return false
							}
							nodeDecided[node] = state.decided.val
							if globalState.Correct[node] {
								decidedValues[state.decided.val] = true
							}
							proposedValues[state.proposed.val] = true
						}
					}
				}
				// C1: Check that all correct nodes decided on some value
				for node, correct := range globalState.Correct {
					if correct {
						if _, ok := nodeDecided[node]; !ok {
							return false
						}
					}
				}
				// Check that at most one value was decided
				if len(decidedValues) > 1 {
					return false
				}
				// Check that a decided values was at some point proposed
				for decidedVal := range decidedValues {
					if !proposedValues[decidedVal] {
						return false
					}
				}
				return true
			},
		),
	)
	resp := checker.Check(sm.StateRoot)
	if ok, out := resp.Response(); !ok {
		t.Errorf("Expected no errors while checking. Got: %v", out)
	}
}
