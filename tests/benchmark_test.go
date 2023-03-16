package gomc_test

import (
	"gomc"
	"gomc/eventManager"
	"gomc/scheduler"
	"testing"
)

type DeliverMsg struct {
	From    int
	To      int
	Message []byte
}

type AckMsg struct {
	From    int
	To      int
	Message []byte
}

type BroadcastNode struct {
	Id        int
	nodes     []int
	Delivered int
	Acked     int
	send      func(int, string, ...any)
}

func (n *BroadcastNode) Broadcast(message []byte) {
	for _, id := range n.nodes {
		n.send(
			id,
			"Deliver",
			message,
		)
	}
}

func (n *BroadcastNode) Deliver(message []byte) {
	n.Delivered++
	for _, id := range n.nodes {
		n.send(
			id,
			"Ack",
			message,
		)
	}
}

func (n *BroadcastNode) Ack(message []byte) {
	n.Acked++
}

type BroadcastState struct {
	delivered int
	acked     int
}

func Benchmark(b *testing.B) {
	numNodes := 2
	for i := 0; i < b.N; i++ {
		sch := scheduler.NewBasicScheduler()
		sm := gomc.NewTreeStateManager(
			func(node *BroadcastNode) BroadcastState {
				return BroadcastState{
					delivered: node.Delivered,
					acked:     node.Acked,
				}
			},
			func(s1, s2 BroadcastState) bool {
				return s1 == s2
			},
		)
		tester := gomc.NewSimulator[BroadcastNode, BroadcastState](sch, 10000, 1000)
		sender := eventManager.NewSender(sch)
		err := tester.Simulate(
			sm,
			func() map[int]*BroadcastNode {
				nodes := map[int]*BroadcastNode{}
				nodeIds := []int{}
				for i := 0; i < numNodes; i++ {
					nodeIds = append(nodeIds, i)
				}
				for _, id := range nodeIds {
					nodes[id] = &BroadcastNode{
						Id:        id,
						send:      sender.SendFunc(id),
						Delivered: 0,
						Acked:     0,
						nodes:     nodeIds,
					}
				}
				return nodes
			},
			[]int{},
			func(t *BroadcastNode) {},
			gomc.NewRequest(0, "Broadcast", []byte("Test Message")),
		)
		if err != nil {
			b.Fatalf("Error while running simulation: %v", err)
		}
	}
}
