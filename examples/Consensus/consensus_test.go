package main

import (
	"bytes"
	"encoding/json"
	"gomc"
	"gomc/checking"
	"gomc/event"
	"gomc/eventManager"
	"os"
	"testing"

	"golang.org/x/exp/slices"
)

type state struct {
	proposed Value[int]
	decided  []Value[int]
}

var predicates = []checking.Predicate[state]{
	checking.Eventually(
		// C1: Termination
		func(s checking.State[state]) bool {
			return checking.ForAllNodes(func(s state) bool {
				return len(s.decided) > 0
			}, s, true)
		},
	),
	func(s checking.State[state]) bool {
		// C2: Validity
		proposed := make(map[Value[int]]bool)
		for _, node := range s.LocalStates {
			proposed[node.proposed] = true
		}
		return checking.ForAllNodes(func(s state) bool {
			if len(s.decided) < 1 {
				// The process has not decided a value yet
				return true
			}
			return proposed[s.decided[0]]
		}, s, false)
	},
	func(s checking.State[state]) bool {
		// C3: Integrity
		return checking.ForAllNodes(func(s state) bool { return len(s.decided) < 2 }, s, false)
	},
	func(s checking.State[state]) bool {
		// C4: Agreement
		decided := make(map[Value[int]]bool)
		checking.ForAllNodes(func(s state) bool {
			for _, val := range s.decided {
				decided[val] = true
			}
			return true
		}, s, true)
		return len(decided) <= 1
	},
}

func TestConsensus(t *testing.T) {
	sim := gomc.Prepare[HierarchicalConsensus[int], state](
		gomc.RandomWalkScheduler(1),
		gomc.MaxRuns(1000),
		gomc.WithPerfectFailureManager(
			func(n *HierarchicalConsensus[int]) { n.crashed = true }, 3, 5,
		),
	)

	nodeIds := []int{1, 2, 3, 4, 5}
	resp := sim.RunSimulation(
		gomc.InitSingleNode(nodeIds,
			func(id int, sp gomc.SimulationParameters) *HierarchicalConsensus[int] {
				send := eventManager.NewSender(sp.Sch)
				node := NewHierarchicalConsensus[int](
					id,
					nodeIds,
					send.SendFunc(id),
				)
				sp.Subscribe(node.Crash)
				return node
			},
		),
		gomc.WithRequests(
			gomc.NewRequest(1, "Propose", Value[int]{2}),
			gomc.NewRequest(2, "Propose", Value[int]{3}),
			gomc.NewRequest(3, "Propose", Value[int]{4}),
			gomc.NewRequest(4, "Propose", Value[int]{5}),
			gomc.NewRequest(5, "Propose", Value[int]{6}),
		),
		gomc.WithTreeStateManager(
			func(node *HierarchicalConsensus[int]) state {
				decided := make([]Value[int], len(node.DecidedVal))
				copy(decided, node.DecidedVal)
				return state{
					proposed: node.ProposedVal,
					decided:  decided,
				}
			},
			func(a, b state) bool {
				if a.proposed.val != b.proposed.val {
					return false
				}
				return slices.Equal(a.decided, b.decided)
			},
		),
		gomc.WithPredicate(predicates...),
		gomc.Export(os.Stdout),
	)
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
	var run []event.EventId
	json.NewDecoder(buffer).Decode(&run)

	sim := gomc.Prepare[HierarchicalConsensus[int], state](
		gomc.ReplayScheduler(run),
		gomc.WithPerfectFailureManager(
			func(n *HierarchicalConsensus[int]) { n.crashed = true }, 3, 5,
		),
	)

	nodeIds := []int{1, 2, 3, 4, 5}
	resp := sim.RunSimulation(
		gomc.InitSingleNode(nodeIds,
			func(id int, sp gomc.SimulationParameters) *HierarchicalConsensus[int] {
				send := eventManager.NewSender(sp.Sch)
				node := NewHierarchicalConsensus[int](
					id,
					nodeIds,
					send.SendFunc(id),
				)
				sp.Subscribe(node.Crash)
				return node
			},
		),
		gomc.WithRequests(
			gomc.NewRequest(1, "Propose", Value[int]{2}),
			gomc.NewRequest(2, "Propose", Value[int]{3}),
			gomc.NewRequest(3, "Propose", Value[int]{4}),
			gomc.NewRequest(4, "Propose", Value[int]{5}),
			gomc.NewRequest(5, "Propose", Value[int]{6}),
		),
		gomc.WithTreeStateManager(
			func(node *HierarchicalConsensus[int]) state {
				decided := make([]Value[int], len(node.DecidedVal))
				copy(decided, node.DecidedVal)
				return state{
					proposed: node.ProposedVal,
					decided:  decided,
				}
			},
			func(a, b state) bool {
				if a.proposed.val != b.proposed.val {
					return false
				}
				return slices.Equal(a.decided, b.decided)
			},
		),
		gomc.WithPredicate(predicates...),
		gomc.Export(os.Stdout),
	)

	if ok, _ := resp.Response(); ok {
		t.Errorf("Expected errors while checking")
	}
}

func BenchmarkConsensus(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sim := gomc.Prepare[HierarchicalConsensus[int], state](
			gomc.RandomWalkScheduler(1),
		)

		nodeIds := []int{1, 2, 3, 4, 5}
		sim.RunSimulation(
			gomc.InitSingleNode(nodeIds,
				func(id int, sp gomc.SimulationParameters) *HierarchicalConsensus[int] {
					send := eventManager.NewSender(sp.Sch)
					node := NewHierarchicalConsensus[int](
						id,
						nodeIds,
						send.SendFunc(id),
					)
					sp.Subscribe(node.Crash)
					return node
				},
			),
			gomc.WithRequests(
				gomc.NewRequest(1, "Propose", Value[int]{2}),
				gomc.NewRequest(2, "Propose", Value[int]{3}),
				gomc.NewRequest(3, "Propose", Value[int]{4}),
				gomc.NewRequest(4, "Propose", Value[int]{5}),
				gomc.NewRequest(5, "Propose", Value[int]{6}),
			),
			gomc.WithTreeStateManager(
				func(node *HierarchicalConsensus[int]) state {
					decided := make([]Value[int], len(node.DecidedVal))
					copy(decided, node.DecidedVal)
					return state{
						proposed: node.ProposedVal,
						decided:  decided,
					}
				},
				func(a, b state) bool {
					if a.proposed.val != b.proposed.val {
						return false
					}
					return slices.Equal(a.decided, b.decided)
				},
			),
			gomc.WithPredicate(predicates...),
		)
	}
}
