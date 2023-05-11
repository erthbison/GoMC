package main

import (
	"gomc"
	"gomc/eventManager"
	"testing"

	"golang.org/x/exp/slices"
)

func BenchmarkConsensus(b *testing.B) {
	sim := gomc.PrepareSimulation(
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
				if a.proposed != b.proposed {
					return false
				}
				return slices.Equal(a.decided, b.decided)
			},
		),
		gomc.PrefixScheduler(),
	)

	nodeIds := []int{1, 2, 3}
	for i := 0; i < b.N; i++ {
		sim.Run(
			gomc.InitSingleNode(nodeIds,
				func(id int, sp gomc.SimulationParameters) *HierarchicalConsensus[int] {
					send := eventManager.NewSender(sp.EventAdder)
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
				2,
			),
		)
	}
}
