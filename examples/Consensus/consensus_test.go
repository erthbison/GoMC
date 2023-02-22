package main

import (
	"bytes"
	"encoding/json"
	"gomc"
	"gomc/eventManager"
	"gomc/predicate"
	"gomc/scheduler"
	"os"
	"testing"
)

type state struct {
	proposed    Value[int]
	decided     Value[int]
	decidedHere bool
}

func TestConsensus(t *testing.T) {
	sch := scheduler.NewQueueScheduler()
	// sch := scheduler.NewRandomScheduler(10000, 1)
	sm := gomc.NewStateManager(
		func(node *HierarchicalConsensus[int]) state {
			decidedHere := false
			select {
			case <-node.DecidedSignal:
				decidedHere = true
			default:
			}
			return state{
				proposed:    node.ProposedVal,
				decided:     node.DecidedVal,
				decidedHere: decidedHere,
			}
		},
		func(a, b state) bool {
			if a.proposed.val != b.proposed.val {
				return false
			}
			return a.decided.val == b.decided.val
		},
	)
	sender := eventManager.NewSender(sch)
	sim := gomc.NewSimulator[HierarchicalConsensus[int], state](sch, sm, 1000, 1000)
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
		predicate.Eventually(
			// C1: Termination
			func(gs gomc.GlobalState[state], _ bool, _ []gomc.GlobalState[state]) bool {
				return predicate.ForAllNodes(func(s state) bool {
					return s.decided.val != 0
				}, gs, true)
			},
		),
		func(gs gomc.GlobalState[state], _ bool, _ []gomc.GlobalState[state]) bool {
			// C2: Validity
			proposed := make(map[Value[int]]bool)
			for _, node := range gs.LocalStates {
				proposed[node.proposed] = true
			}
			return predicate.ForAllNodes(func(s state) bool { return proposed[s.decided] }, gs, false)
		},
		func(gs gomc.GlobalState[state], _ bool, seq []gomc.GlobalState[state]) bool {
			// C3: Integrity
			numDecided := make(map[int]int)
			for _, state := range seq {
				for id, node := range state.LocalStates {
					if node.decidedHere {
						numDecided[id]++
					}
				}
			}
			for id := range gs.LocalStates {
				if numDecided[id] > 1 {
					return false
				}
			}
			return true
		},
		func(gs gomc.GlobalState[state], _ bool, seq []gomc.GlobalState[state]) bool {
			// C4: Agreement
			decided := make(map[Value[int]]bool)
			predicate.ForAllNodes(func(s state) bool {
				if s.decided.val != 0 {
					decided[s.decided] = true
				}
				return true
			}, gs, true)
			return len(decided) <= 1
		},
	)
	resp := checker.Check(sm.StateRoot)
	if ok, out := resp.Response(); !ok {
		t.Errorf("Expected no errors while checking. Got: %v", out)

		var buffer bytes.Buffer
		json.NewEncoder(&buffer).Encode(resp.Export())
		os.WriteFile("FailedRun.txt", buffer.Bytes(), 0755)
	}
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
			}
		},
		func(a, b state) bool {
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
		predicate.Eventually(
			// C1: Termination
			func(gs gomc.GlobalState[state], _ bool, _ []gomc.GlobalState[state]) bool {
				return predicate.ForAllNodes(func(s state) bool {
					return s.decided.val != 0
				}, gs, true)
			},
		),
		func(gs gomc.GlobalState[state], _ bool, _ []gomc.GlobalState[state]) bool {
			// C2: Validity
			proposed := make(map[Value[int]]bool)
			for _, node := range gs.LocalStates {
				proposed[node.proposed] = true
			}
			return predicate.ForAllNodes(func(s state) bool { return proposed[s.decided] }, gs, false)
		},
		func(gs gomc.GlobalState[state], _ bool, seq []gomc.GlobalState[state]) bool {
			// C3: Integrity
			numDecided := make(map[int]int)
			for _, state := range seq {
				for id, node := range state.LocalStates {
					if node.decidedHere {
						numDecided[id]++
					}
				}
			}
			for id := range gs.LocalStates {
				if numDecided[id] > 1 {
					return false
				}
			}
			return true
		},
		func(gs gomc.GlobalState[state], _ bool, seq []gomc.GlobalState[state]) bool {
			// C4: Agreement
			decided := make(map[Value[int]]bool)
			predicate.ForAllNodes(func(s state) bool {
				if s.decided.val != 0 {
					decided[s.decided] = true
				}
				return true
			}, gs, true)
			return len(decided) <= 1
		},
	)
	resp := checker.Check(sm.StateRoot)
	if ok, _ := resp.Response(); ok {
		t.Errorf("Expected errors while checking.")
	}
}
