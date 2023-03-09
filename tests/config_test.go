package gomc_test

import (
	"fmt"
	"gomc"
	"testing"
)

func TestConfig(t *testing.T) {
	sim := gomc.Prepare[BroadcastNode, BroadcastState](
		gomc.QueueScheduler(),
		gomc.MaxDepth(10000),
	)
	resp := sim.RunSimulation(
		gomc.InitNodeFunc(
			func() map[int]*BroadcastNode {
				nodes := map[int]*BroadcastNode{}
				nodeIds := []int{}
				for i := 0; i < 2; i++ {
					nodeIds = append(nodeIds, i)
				}
				for _, id := range nodeIds {
					nodes[id] = &BroadcastNode{
						Id:        id,
						send:      sim.SendFactory(id),
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

	// nodeIds := []int{0, 1, 2}
	// resp = sim.RunSimulation(
	// 	gomc.InitSingleNode(
	// 		nodeIds,
	// 		func(id int) *BroadcastNode {
	// 			return &BroadcastNode{
	// 				Id:        id,
	// 				send:      sim.SendFactory(id),
	// 				Delivered: 0,
	// 				Acked:     0,
	// 				nodes:     nodeIds,
	// 			}
	// 		},
	// 	),
	// 	gomc.WithRequests(
	// 		gomc.NewRequest(0, "Broadcast", []byte("Test Message")),
	// 	),
	// 	gomc.WithPredicate[State](),
	// )

	// if ok, txt := resp.Response(); !ok {
	// 	fmt.Println(txt)
	// 	resp.Export()
	// }
}
