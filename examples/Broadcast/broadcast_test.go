package main

import (
	"fmt"
	"gomc"
	"testing"
)

type State struct {
	delivered int
	acked     int
}

func TestBroadcast(t *testing.T) {
	numNodes := 2
	sim := gomc.Prepare[Node, State](
		gomc.QueueScheduler(),
	)
	resp := sim.RunSimulation(
		gomc.InitNodeFunc(func() map[int]*Node {
			nodeMap := map[int]*Node{}
			nodes := []int{}
			for i := 0; i < numNodes; i++ {
				nodes = append(nodes, i)
			}
			for _, id := range nodes {
				nodeMap[id] = &Node{
					Id:        id,
					send:      sim.SendFactory(id),
					Delivered: 0,
					Acked:     0,
					nodes:     nodes,
					sleep:     sim.SleepFactory(id),
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
	)
	if ok, out := resp.Response(); !ok {
		fmt.Print(out)
	}
}
