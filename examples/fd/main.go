package main

import (
	"fmt"
	"gomc"
	"gomc/scheduler"
	"time"

	"golang.org/x/exp/slices"
)

type State struct {
	crashed []int
}

func main() {
	numNodes := 3
	sch := scheduler.NewRandomScheduler(500)
	sm := gomc.NewStateManager(
		func(t *fd) State {
			crashed := make([]int, len(t.crashed))
			for i, val := range t.crashed {
				crashed[i] = val
			}
			return State{
				crashed: crashed,
			}
		},
		func(s1, s2 State) bool {
			return slices.Equal(s1.crashed, s2.crashed)
		},
	)
	tester := gomc.NewSimulator[fd, State](sch, sm)
	sender := gomc.NewSender(sch)
	sleep := gomc.NewSleepManager(sch, tester.NextEvt)
	err := tester.Simulate(
		func() map[int]*fd {
			ids := []int{}
			for i := 0; i < numNodes; i++ {
				ids = append(ids, i)
			}
			nodes := map[int]*fd{}
			for _, id := range ids {
				nodes[id] = NewFd(
					ids, 5*time.Second, sender.SendFunc(id), sleep.SleepFunc(id),
				)
			}
			return nodes
		},
		[]int{2},
		gomc.NewFunc(0, "Start"),
	)

	if err != nil {
		panic(err)
	}
	// fmt.Println(sch.EventRoot.Newick())
	fmt.Println(sm.StateRoot.Newick())
}