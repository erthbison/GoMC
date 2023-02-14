package main

import (
	"fmt"
	"gomc"
	"gomc/scheduler"
	"testing"
)

type State struct {
	delivered int
	acked     int
}

func TestBroadcast(t *testing.T) {
	numNodes := 2
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
	sleep := gomc.NewSleepManager(sch, tester.NextEvt)
	sender := gomc.NewSender(sch)
	err := tester.Simulate(
		func() map[int]*Node {
			nodeMap := map[int]*Node{}
			nodes := []int{}
			for i := 0; i < numNodes; i++ {
				nodes = append(nodes, i)
			}
			for _, id := range nodes {
				nodeMap[id] = &Node{
					Id:        id,
					send:      sender.SendFunc(id),
					Delivered: 0,
					Acked:     0,
					nodes:     nodes,
					sleep:     sleep.SleepFunc(id),
				}
			}
			return nodeMap
		},
		[]int{},
		gomc.NewRequest(0, "Broadcast", []byte("0")),
	)
	if err != nil {
		t.Errorf("Expected no error")
	}
	fmt.Println(sch.EventRoot)
	fmt.Println(sm.StateRoot)

	fmt.Println(sch.EventRoot.Newick())
	fmt.Println(sm.StateRoot.Newick())

}
