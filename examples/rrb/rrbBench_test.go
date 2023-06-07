package main

import (
	"gomc"
	"gomc/eventManager"
	"testing"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func BenchmarkRrb(b *testing.B) {
	nodeIds := []int{0, 1, 2}
	crashedNodes := []int{1}

	sim := gomc.PrepareSimulation(
		gomc.WithTreeStateManager(
			func(node *Rrb) State {
				return State{
					delivered:      maps.Clone(node.delivered),
					sent:           maps.Clone(node.sent),
					deliveredSlice: slices.Clone(node.deliveredSlice),
				}
			},
			func(s1, s2 State) bool {
				if !maps.Equal(s1.delivered, s2.delivered) {
					return false
				}
				if !slices.Equal(s1.deliveredSlice, s2.deliveredSlice) {
					return false
				}
				return maps.Equal(s1.sent, s2.sent)
			},
		),
		gomc.PrefixScheduler(),
	)

	for i := 0; i < b.N; i++ {
		sim.Run(
			gomc.InitNodeFunc(
				func(sp eventManager.SimulationParameters) map[int]*Rrb {
					send := eventManager.NewSender(sp)
					nodes := map[int]*Rrb{}
					for _, id := range nodeIds {
						nodes[id] = NewRrb(
							id,
							nodeIds,
							send.SendFunc(id),
						)
					}
					return nodes
				},
			),
			gomc.WithRequests(
				gomc.NewRequest(0, "Broadcast", "Test Message"),
			),
			gomc.WithPredicateChecker(predicates...),
			gomc.WithPerfectFailureManager(
				func(t *Rrb) { t.crashed = true },
				crashedNodes...,
			),
		)
	}
}
