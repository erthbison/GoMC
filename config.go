package gomc

import (
	"fmt"
	"gomc/scheduler"
	"log"
)

// A struct used to quickly configure the simulation
type Config[T, S any] struct {
	// if the value "random" a RandomScheduler is used. Otherwise a BasicScheduler is used
	Scheduler string
	// Number of runs to be used with a RandomScheduler. Default value is 10000. If value is 0 default value is used.
	NumRuns uint

	GetLocalState func(*T) S
	StatesEqual   func(S, S) bool

	InitNodes func() map[int]*T

	StartFuncs     map[int][]func(*T) error
	IncorrectNodes []int

	Preds []func(GlobalState[S], bool, []GlobalState[S]) bool
}

func RunSimulation[T, S any](cfg Config[T, S]) {
	var sch scheduler.Scheduler[T]
	if cfg.Scheduler == "random" {
		if cfg.NumRuns == 0 {
			// If numRuns is 0 set to default value 10 000
			cfg.NumRuns = 10000
		}
		sch = scheduler.NewRandomScheduler[T](cfg.NumRuns)
	} else {
		sch = scheduler.NewBasicScheduler[T]()
	}

	sm := NewStateManager(
		cfg.GetLocalState,
		cfg.StatesEqual,
	)

	// If incorrectNodes is not provided use an empty slice
	if cfg.IncorrectNodes == nil {
		cfg.IncorrectNodes = make([]int, 0)
	}

	sim := NewSimulator[T, S](sch, sm)
	err := sim.Simulate(cfg.InitNodes, cfg.StartFuncs, cfg.IncorrectNodes)
	if err != nil {
		log.Panicf("Received an error while running simulation: %v", err)
	}

	checker := NewPredicateChecker(cfg.Preds...)
	resp := checker.Check(sm.StateRoot)
	_, text := resp.Response()
	fmt.Println(text)

}
