package main

import (
	"bytes"
	"encoding/json"
	"gomc"
	"gomc/event"
	"gomc/eventManager"
	"os"
	"testing"

	"golang.org/x/exp/slices"
)

func TestConsensusReplay(t *testing.T) {
	in, err := os.ReadFile("FailedRun.txt")
	if err != nil {
		t.Errorf("Error while setting up test: %v", err)
	}
	buffer := bytes.NewBuffer(in)
	var run []event.EventId
	json.NewDecoder(buffer).Decode(&run)

	sim := gomc.PrepareSimulation(
		gomc.WithTreeStateManager(
			func(node *HierarchicalConsensus[int]) state {
				decided := make([]Value[int], len(node.DecidedVal))
				copy(decided, node.DecidedVal)
				return state{
					proposed: node.ProposedVal,
					decided:  decided,
				}
			},
			func(a, b state) bool {
				if a.proposed.Val != b.proposed.Val {
					return false
				}
				return slices.Equal(a.decided, b.decided)
			},
		),
		gomc.ReplayScheduler(run),
	)

	nodeIds := []int{1, 2, 3}
	resp := sim.Run(
		gomc.InitSingleNode(nodeIds,
			func(id int, sp gomc.SimulationParameters) *HierarchicalConsensus[int] {
				send := eventManager.NewSender(sp.EventAdder)
				node := NewHierarchicalConsensus[int](
					id,
					nodeIds,
					send.SendFunc(id),
				)
				sp.CrashSubscribe(id, node.Crash)
				return node
			},
		),
		gomc.WithRequests(
			gomc.NewRequest(1, "Propose", Value[int]{1}),
			gomc.NewRequest(2, "Propose", Value[int]{2}),
			gomc.NewRequest(3, "Propose", Value[int]{3}),
		),
		gomc.WithPredicateChecker(predicates...),
		gomc.WithPerfectFailureManager(
			func(t *HierarchicalConsensus[int]) { t.crashed = true },
			1,
		),
		gomc.Export(os.Stdout),
	)

	if ok, _ := resp.Response(); ok {
		t.Errorf("Expected errors while checking")
	}
}
