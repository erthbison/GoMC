package gomc_test

import (
	"fmt"
	"gomc"
	"os"
	"testing"
)

func TestConfig(t *testing.T) {
	sim := gomc.ConfigureSimulation(
		gomc.SimulationConfig[BroadcastNode, BroadcastState]{
			Scheduler: "random",
			NumRuns:   10000,
			MaxDepth:  1000,
			Seed:      1,
			GetLocalState: func(node *BroadcastNode) BroadcastState {
				return BroadcastState{
					delivered: node.Delivered,
					acked:     node.Acked,
				}
			},
			StatesEqual: func(s1, s2 BroadcastState) bool {
				return s1 == s2
			},
		},
	)
	resp := sim.RunSimulation(
		func() map[int]*BroadcastNode {
			nodes := map[int]*BroadcastNode{}
			for i := 0; i < 2; i++ {
				nodes[i] = &BroadcastNode{
					Id:        i,
					send:      sim.SendFactory(i),
					Delivered: 0,
					Acked:     0,
					nodes:     map[int]*BroadcastNode{},
				}
			}
			for _, node := range nodes {
				node.RegisterNodes(nodes)
			}
			return nodes
		},
		gomc.NewRequest(0, "Broadcast", []byte("Test Message")),
	)
	if ok, txt := resp.Response(); !ok {
		fmt.Println(txt)
		resp.Export()
	}

	sim.WriteStateTree(os.Stdout)
}
