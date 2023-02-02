package main

import (
	"experimentation/tester"
	"fmt"
)

type State string

func main() {
	numNodes := 2
	sch := tester.NewBasicScheduler()
	sm := tester.NewStateManager(
		func(node *fifo) State {
			return State(fmt.Sprintf("%v", len(node.Received)))
		},
		func(s1, s2 State) bool {
			return s1 == s2
		},
	)
	tester := tester.CreateSimulator[fifo, State](sch, sm)
	tester.Simulate(
		func() map[int]*fifo {
			nodes := map[int]*fifo{}
			for i := 0; i < numNodes; i++ {
				nodes[i] = &fifo{
					send: tester.Send,
				}
			}
			return nodes
		},
		func(nodes map[int]*fifo) {
			for i := 0; i < 3; i++ {
				nodes[0].Send(1, []byte(fmt.Sprintf("Test Message - %v", i)))
			}
		},
	)
	fmt.Println(sch.EventRoot)
	fmt.Println(sm.StateRoot)

	fmt.Println(sm.StateRoot.Newick())
}
