package main

import (
	"experimentation/gomc"
	"fmt"
)

type State string

func main() {
	numNodes := 2
	sch := gomc.NewBasicScheduler[fifo]()
	sm := gomc.NewStateManager(
		func(node *fifo) State {
			return State(fmt.Sprintf("%v", len(node.Received)))
		},
		func(s1, s2 State) bool {
			return s1 == s2
		},
	)
	tester := gomc.NewSimulator[fifo, State](sch, sm)
	err := tester.Simulate(
		func() map[int]*fifo {
			nodes := map[int]*fifo{}
			for i := 0; i < numNodes; i++ {
				nodes[i] = &fifo{
					send: tester.Send,
				}
			}
			return nodes
		},
		func(nodes map[int]*fifo) error {
			for i := 0; i < 3; i++ {
				nodes[0].Send(1, []byte(fmt.Sprintf("Test Message - %v", i)))
			}
			return nil
		},
	)

	if err != nil {
		panic(err)
	}
	fmt.Println(sch.EventRoot)
	fmt.Println(sm.StateRoot)

	fmt.Println(sm.StateRoot.Newick())
}
