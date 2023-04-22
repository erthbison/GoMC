package main

import (
	"fmt"
	"gomc"
	"gomc/eventManager"
	"testing"
)

type State string

func TestFifo(t *testing.T) {
	numNodes := 2

	sim := gomc.Prepare[fifo, State](
		gomc.PrefixScheduler(),
	)
	sim.RunSimulation(
		gomc.InitNodeFunc(
			func(sp gomc.SimulationParameters) map[int]*fifo {
				nodes := map[int]*fifo{}
				send := eventManager.NewSender(sp.EventAdder)
				for i := 0; i < numNodes; i++ {
					nodes[i] = &fifo{
						send: send.SendFunc(i),
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
