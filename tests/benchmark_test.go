package gomc_test

import (
	"gomc"
	"gomc/eventManager"
	"testing"
)

func Benchmark(b *testing.B) {
	numNodes := 5
	sim := gomc.PrepareSimulation(
		gomc.WithTreeStateManager(
			func(node *BroadcastNode) BroadcastState {
				return BroadcastState{
					delivered: node.Delivered,
					acked:     node.Acked,
				}
			},
			func(s1, s2 BroadcastState) bool {
				return s1 == s2
			},
		),
		gomc.PrefixScheduler(),
	)
	for i := 0; i < b.N; i++ {
		sim.Run(
			gomc.InitNodeFunc(
				func(sp eventManager.SimulationParameters) map[int]*BroadcastNode {
					send := eventManager.NewSender(sp)
					nodes := map[int]*BroadcastNode{}
					nodeIds := []int{}
					for i := 0; i < numNodes; i++ {
						nodeIds = append(nodeIds, i)
					}
					for _, id := range nodeIds {
						nodes[id] = &BroadcastNode{
							Id:        id,
							send:      send.SendFunc(id),
							Delivered: 0,
							Acked:     0,
							nodes:     nodeIds,
						}
					}
					return nodes
				},
			),
			gomc.WithRequests(gomc.NewRequest(0, "Broadcast", []byte("Test Message"))),
			gomc.WithPredicateChecker[BroadcastState](),
		)
	}
}
