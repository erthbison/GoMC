package main

import (
	"fmt"
	"gomc"
	"gomc/eventManager"
	"testing"
)

type State struct {
	delivered int
	acked     int
}

func TestBroadcast(t *testing.T) {
	numNodes := 2
	sim := gomc.PrepareSimulation[Node, State](
		gomc.PrefixScheduler(),
	)
	resp := sim.Run(
		gomc.InitNodeFunc(func(sp gomc.SimulationParameters) map[int]*Node {
			send := eventManager.NewSender(sp.EventAdder)
			sleep := eventManager.NewSleepManager(sp.EventAdder, sp.NextEvt)
			nodeMap := map[int]*Node{}
			nodes := []int{}
			for i := 0; i < numNodes; i++ {
				nodes = append(nodes, i)
			}
			for _, id := range nodes {
				nodeMap[id] = &Node{
					Id:        id,
					send:      send.SendFunc(id),
					Delivered: 0,
					Acked:     0,
					nodes:     nodes,
					sleep:     sleep.SleepFunc(id),
				}
			}
			return nodeMap
		}),
		gomc.WithRequests(
			gomc.NewRequest(0, "Broadcast", []byte("0")),
		),
		gomc.WithTreeStateManager(
			func(node *Node) State {
				return State{
					delivered: node.Delivered,
					acked:     node.Acked,
				}
			},
			func(s1, s2 State) bool {
				return s1 == s2
			},
		),
		gomc.WithPredicateChecker[State](),
	)
	if ok, out := resp.Response(); !ok {
		fmt.Print(out)
	}
}
