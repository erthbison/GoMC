package main

import (
	"fmt"
	"gomc"
	"gomc/scheduler"
)

type State struct {
	delivered int
	acked     int
}

func main() {
	numNodes := 2
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
	sleep := gomc.NewSleepManager[Node](sch, tester.NextEvt)
	sender := gomc.NewSender[Node](sch)
	err := tester.Simulate(func() map[int]*Node {
		nodeMap := map[int]*Node{}
		nodes := []int{}
		for i := 0; i < numNodes; i++ {
			nodes = append(nodes, i)
		}
		for _, id := range nodes {
			nodeMap[id] = &Node{
				Id:        id,
				send:      sender.Send,
				Delivered: 0,
				Acked:     0,
				nodes:     nodes,
				sleep:     sleep.SleepFunc(id),
			}
		}
		return nodeMap
	},
		func(nodes map[int]*Node) error {
			nodes[0].Broadcast([]byte("1"))
			return nil
		},
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(sch.EventRoot)
	fmt.Println(sm.StateRoot)

	fmt.Println(sch.EventRoot.Newick())
	fmt.Println(sm.StateRoot.Newick())

}
