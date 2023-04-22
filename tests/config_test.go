package gomc_test

import (
	"fmt"
	"gomc"
	"gomc/eventManager"
	"testing"
)

func TestConfig(t *testing.T) {
	sim := gomc.Prepare[BroadcastNode, BroadcastState](
		gomc.PrefixScheduler(),
		gomc.MaxDepth(10000),
	)
	resp := sim.RunSimulation(
		gomc.InitNodeFunc(
			func(sp gomc.SimulationParameters) map[int]*BroadcastNode {
				send := eventManager.NewSender(sp.EventAdder)
				nodes := map[int]*BroadcastNode{}
				nodeIds := []int{}
				for i := 0; i < 2; i++ {
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
		gomc.WithRequests(
			gomc.NewRequest(0, "Broadcast", []byte("Test Message")),
		),
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
	)

	if ok, txt := resp.Response(); !ok {
		fmt.Println(txt)
		resp.Export()
	}

}
