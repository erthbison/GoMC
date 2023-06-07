package main

import (
	"gomc"
	"gomc/eventManager"
	"testing"

	"golang.org/x/exp/slices"
)

func BenchmarkOnrr(b *testing.B) {

	nodeIds := []int{1, 2, 3}

	sim := gomc.PrepareSimulation(
		gomc.WithTreeStateManager(
			func(node *onrr) State {
				reads := []int{}
				reads = append(reads, node.possibleReads...)

				// If there has been a read indication store it. Otherwise ignore it
				read := 0
				currentRead := false
				select {
				case read = <-node.ReadIndicator:
					currentRead = true
				default:
				}

				return State{
					ongoingRead:   node.ongoingRead,
					ongoingWrite:  node.ongoingWrite,
					possibleReads: reads,
					read:          read,
					currentRead:   currentRead,
				}
			},
			func(a, b State) bool {
				if a.ongoingRead != b.ongoingRead {
					return false
				}
				if a.ongoingWrite != b.ongoingWrite {
					return false
				}
				if a.currentRead != b.currentRead {
					return false
				}
				if a.read != b.read {
					return false
				}
				return slices.Equal(a.possibleReads, b.possibleReads)
			},
		),
		gomc.RandomWalkScheduler(1),
		gomc.MaxRuns(10000),
	)

	for i := 0; i < b.N; i++ {
		sim.Run(
			gomc.InitNodeFunc(
				func(sp eventManager.SimulationParameters) map[int]*onrr {
					send := eventManager.NewSender(sp)

					nodes := make(map[int]*onrr)
					for _, id := range nodeIds {
						nodes[id] = NewOnrr(id, send.SendFunc(id), nodeIds)
					}
					go func() {
						for {
							<-nodes[1].WriteIndicator
						}
					}()
					return nodes
				},
			),
			gomc.WithRequests(
				gomc.NewRequest(1, "Write", 2),
				gomc.NewRequest(2, "Read"),
				gomc.NewRequest(3, "Read"),
			),
			gomc.WithPredicateChecker(predicates...),
		)
	}
}
