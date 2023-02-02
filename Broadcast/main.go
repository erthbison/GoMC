package main

import (
	"experimentation/tester"
	"fmt"
)

type State string

func main() {
	numNodes := 2
	// sch := tester.NewRunScheduler()
	sch := tester.NewBasicScheduler()
	sm := tester.NewStateManager(
		func(node *Node) State {
			return State(fmt.Sprintf("%v%v", node.Delivered, node.Acked))
		},
		func(s1, s2 State) bool {
			return s1 == s2
		},
	)
	tester := tester.CreateSimulator[Node, State](sch, sm)
	tester.Simulate(func() map[int]*Node {
		nodeMap := map[int]*Node{}
		nodes := []int{}
		for i := 0; i < numNodes; i++ {
			nodes = append(nodes, i)
		}
		for _, id := range nodes {
			nodeMap[id] = &Node{
				Id:        id,
				send:      tester.Send,
				Delivered: 0,
				Acked:     0,
				nodes:     nodes,
			}
		}
		return nodeMap
	},
		func(nodes map[int]*Node) {
			nodes[0].Broadcast([]byte("1"))
		},
	)
	fmt.Println(sch.EventRoot)
	fmt.Println(sm.StateRoot)

	fmt.Println(sch.EventRoot.Newick())
	fmt.Println(sm.StateRoot.Newick())

}
