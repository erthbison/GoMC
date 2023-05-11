package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gomc"
	"gomc/event"
	"os"
	"testing"
)

func TestGrpcConsensusCreateReplay(t *testing.T) {
	sim := gomc.PrepareSimulation(
		gomc.WithTreeStateManager(getState, cmpState),
		gomc.PrefixScheduler(),
	)

	resp := sim.Run(
		gomc.InitNodeFunc(createNodes(addrMap)),
		gomc.WithRequests(
			gomc.NewRequest(1, "Propose", "1"),
			gomc.NewRequest(2, "Propose", "2"),
			gomc.NewRequest(3, "Propose", "3"),
		),
		gomc.WithPredicateChecker(predicates...),
		gomc.WithPerfectFailureManager(
			func(t *GrpcConsensus) { t.Stop() }, 1,
		),
		gomc.WithStopFunction(func(t *GrpcConsensus) { t.Stop() }),
	)

	fmt.Println()
	ok, text := resp.Response()
	if ok {
		t.Errorf("Expected simulation to fail. Got:\n %v", text)
	} else {
		var buffer bytes.Buffer
		json.NewEncoder(&buffer).Encode(resp.Export())
		os.WriteFile("FailedRun.txt", buffer.Bytes(), 0755)
	}
}

func TestReplayConsensus(t *testing.T) {
	in, err := os.ReadFile("FailedRun.txt")
	if err != nil {
		t.Errorf("Error while setting up test: %v", err)
	}
	buffer := bytes.NewBuffer(in)
	var run []event.EventId
	json.NewDecoder(buffer).Decode(&run)

	sim := gomc.PrepareSimulation(
		gomc.WithTreeStateManager(getState, cmpState),
		gomc.ReplayScheduler(run),
	)

	resp := sim.Run(
		gomc.InitNodeFunc(
			createNodes(addrMap),
		),
		gomc.WithRequests(
			gomc.NewRequest(1, "Propose", "1"),
			gomc.NewRequest(2, "Propose", "2"),
			gomc.NewRequest(3, "Propose", "3"),
		),
		gomc.WithPredicateChecker(predicates...),
		gomc.WithPerfectFailureManager(
			func(t *GrpcConsensus) { t.Stop() }, 1,
		),
		gomc.WithStopFunction(func(t *GrpcConsensus) { t.Stop() }),
	)

	fmt.Println()
	ok, text := resp.Response()
	if ok {
		t.Errorf("Expected simulation to fail. Got:\n %v", text)
	}
}
