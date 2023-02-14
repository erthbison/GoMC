package gomc_test

import (
	"gomc"
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

type Node struct {
	Id        int
	nodes     map[int]*Node
	Delivered int
	Acked     int
	send      func(int, string, []byte)
}

func (n *Node) RegisterNodes(nodes map[int]*Node) {
	for id, node := range nodes {
		n.nodes[id] = node
	}
}

func (n *Node) Broadcast(message []byte) {
	for id := range n.nodes {
		n.send(
			id,
			"Deliver",
			message,
		)
	}
}

func (n *Node) Deliver(from int, message []byte) {
	n.Delivered++
	for id := range n.nodes {
		n.send(
			id,
			"Ack",
			message,
		)
	}
}

func (n *Node) Ack(from int, message []byte) {
	n.Acked++
}

type State struct {
	delivered int
	acked     int
}

func Benchmark(b *testing.B) {
	numNodes := 2
	for i := 0; i < b.N; i++ {
		sch := scheduler.NewBasicScheduler()
		sm := gomc.NewStateManager(
			func(node *Node) State {
				return State{
					delivered: node.Delivered,
					acked:     node.Acked,
				}
			},
			func(s1, s2 State) bool {
				return s1 == s2
			},
		)
		tester := gomc.NewSimulator[Node, State](sch, sm)
		sender := gomc.NewSender(sch)
		err := tester.Simulate(
			func() map[int]*Node {
				nodes := map[int]*Node{}
				for i := 0; i < numNodes; i++ {
					nodes[i] = &Node{
						Id:        i,
						send:      sender.SendFunc(i),
						Delivered: 0,
						Acked:     0,
						nodes:     map[int]*Node{},
					}
				}
				for _, node := range nodes {
					node.RegisterNodes(nodes)
				}
				return nodes
			},
			[]int{},
			gomc.NewFunc(0, "Broadcast", []byte("Test Message")),
		)
		if err != nil {
			b.Fatalf("Error while running simulation: %v", err)
		}
	}
}
