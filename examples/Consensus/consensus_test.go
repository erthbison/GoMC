package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"golang.org/x/exp/slices"

	"gomc"
	"gomc/checking"
	"gomc/eventManager"
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
	sim := gomc.PrepareSimulation(
		gomc.WithTreeStateManager(
			func(node *HierarchicalConsensus[int]) state {
				return state{
					proposed: node.ProposedVal,
					decided:  slices.Clone(node.DecidedVal),
				}
			},
			func(a, b state) bool {
				if a.proposed != b.proposed {
					return false
				}
				return slices.Equal(a.decided, b.decided)
			},
		),
		gomc.PrefixScheduler(),
	)

	nodeIds := []int{1, 2, 3}
	resp := sim.Run(
		gomc.InitSingleNode(nodeIds,
			func(id int, sp eventManager.SimulationParameters) *HierarchicalConsensus[int] {
				send := eventManager.NewSender(sp)
				node := NewHierarchicalConsensus[int](
					id,
					nodeIds,
					send.SendFunc(id),
				)
				sp.CrashSubscribe(id, node.Crash)
				return node
			},
		),
		gomc.WithRequests(
			gomc.NewRequest(1, "Propose", Value[int]{1}),
			gomc.NewRequest(2, "Propose", Value[int]{2}),
			gomc.NewRequest(3, "Propose", Value[int]{3}),
		),
		gomc.WithPredicateChecker(predicates...),
		gomc.WithPerfectFailureManager(
			func(t *HierarchicalConsensus[int]) { t.crashed = true },
			1,
		),
		gomc.Export(os.Stdout),
	)
	if ok, out := resp.Response(); !ok {
		t.Errorf("Expected no errors while checking. Got: %v", out)

		var buffer bytes.Buffer
		json.NewEncoder(&buffer).Encode(resp.Export())
		os.WriteFile("FailedRun.txt", buffer.Bytes(), 0755)
	}
}
