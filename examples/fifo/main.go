package main

import (
	"fmt"
	"gomc"
	"gomc/scheduler"
)

type State string

func main() {
	numNodes := 2
	sch := scheduler.NewBasicScheduler[fifo]()
	sm := gomc.NewStateManager(
		func(node *fifo) State {
			return State(fmt.Sprintf("%v", len(node.Received)))
		},
		func(s1, s2 State) bool {
			return s1 == s2
		},
	)
	tester := gomc.NewSimulator[fifo, State](sch, sm)
	sender := gomc.NewSender[fifo](sch)
	err := tester.Simulate(
		func() map[int]*fifo {
			nodes := map[int]*fifo{}
			for i := 0; i < numNodes; i++ {
				nodes[i] = &fifo{
					send: sender.SendFunc(i),
				}
			}
			return nodes
		},
		map[int][]func(*fifo) error{
			0: {
				func(node *fifo) error {
					for i := 0; i < 3; i++ {
						node.Send(1, []byte(fmt.Sprintf("Test Message - %v", i)))
					}
					return nil
				},
			},
		},
		[]int{},
	)

	if err != nil {
		panic(err)
	}
	fmt.Println(sch.EventRoot)
	fmt.Println(sm.StateRoot)

	fmt.Println(sm.StateRoot.Newick())
}
