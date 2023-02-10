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
	send      func(int, int, string, []byte)
}

func (n *Node) RegisterNodes(nodes map[int]*Node) {
	for id, node := range nodes {
		n.nodes[id] = node
	}
}

func (n *Node) Broadcast(message []byte) {
	for id := range n.nodes {
		n.send(
			n.Id,
			id,
			"Deliver",
			message,
		)
	}
}

func (n *Node) Deliver(from, to int, message []byte) {
	n.Delivered++
	for id := range n.nodes {
		n.send(
			n.Id,
			id,
			"Ack",
			message,
		)
	}
}

func (n *Node) Ack(from, to int, message []byte) {
	n.Acked++
}

type State struct {
	delivered int
	acked     int
}

func Benchmark(b *testing.B) {
	numNodes := 2
	for i := 0; i < b.N; i++ {
		sch := scheduler.NewBasicScheduler[Node]()
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
		sender := gomc.NewSender[Node](sch)
		err := tester.Simulate(
			func() map[int]*Node {
				nodes := map[int]*Node{}
				for i := 0; i < numNodes; i++ {
					nodes[i] = &Node{
						Id:        i,
						send:      sender.Send,
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
			func(nodes map[int]*Node) error {
				nodes[0].Broadcast([]byte("Test Message"))
				return nil
			},
		)
		if err != nil {
			panic(err)
		}
	}
}
