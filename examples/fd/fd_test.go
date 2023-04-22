package main

import (
	"gomc"
	"gomc/eventManager"
	"testing"
	"time"

	"golang.org/x/exp/slices"
)

type State struct {
	crashed []int
}

func TestFd(t *testing.T) {
	sim := gomc.Prepare[fd, State](
		gomc.RandomWalkScheduler(500),
		gomc.WithPerfectFailureManager(
			func(t *fd) { t.stopped = true }, 2,
		),
	)

	nodeIds := []int{0, 1, 2}
	sim.RunSimulation(
		gomc.InitSingleNode(
			nodeIds,
			func(id int, sp gomc.SimulationParameters) *fd {
				send := eventManager.NewSender(sp.EventAdder)
				sleep := eventManager.NewSleepManager(sp.EventAdder, sp.NextEvt)
				return NewFd(
					id, nodeIds, 5*time.Second, send.SendFunc(id), sleep.SleepFunc(id),
				)
			},
		),
		gomc.WithRequests(gomc.NewRequest(0, "Start")),
		gomc.WithTreeStateManager(
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
		),
	)
}
