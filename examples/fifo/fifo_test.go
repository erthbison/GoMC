package main

import (
	"fmt"
	"gomc"
	"testing"
)

type State string

func TestFifo(t *testing.T) {
	numNodes := 2

	sim := gomc.Prepare[fifo, State](
		gomc.BasicScheduler(),
	)
	sim.RunSimulation(
		gomc.InitNodeFunc(
			func() map[int]*fifo {
				nodes := map[int]*fifo{}
				for i := 0; i < numNodes; i++ {
					nodes[i] = &fifo{
						send: sim.SendFactory(i),
					}
				}
				return nodes
			},
		),
		gomc.WithRequests(
			gomc.NewRequest(0, "Send", 1, []byte("Test Message - 1")),
			gomc.NewRequest(0, "Send", 1, []byte("Test Message - 2")),
			gomc.NewRequest(0, "Send", 1, []byte("Test Message - 3")),
		),
		gomc.WithTreeStateManager(
			func(node *fifo) State {
				return State(fmt.Sprintf("%v", len(node.Received)))
			},
			func(s1, s2 State) bool {
				return s1 == s2
			},
		),
	)
}
