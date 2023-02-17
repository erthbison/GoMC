package gomc

import (
	"fmt"
	"gomc/scheduler"
	"log"
	"time"
)

// A struct used to quickly configure the simulation
type Config[T, S any] struct {
	// if the value "random" a RandomScheduler is used. Otherwise a BasicScheduler is used
	Scheduler string
	// Number of runs to be used with a RandomScheduler. Default value is 10000. If value is 0 default value is used.
	NumRuns uint
	// Maximum depth of a run. Default value is 1000
	MaxDepth uint

	GetLocalState func(*T) S
	StatesEqual   func(S, S) bool
}

func ConfigureSimulation[T, S any](cfg Config[T, S]) SimulationRunner[T, S] {
	if cfg.NumRuns == 0 {
		// If numRuns is 0 set to default value 10 000
		cfg.NumRuns = 10000
	}
	if cfg.MaxDepth == 0 {
		cfg.MaxDepth = 1000
	}
	var sch scheduler.Scheduler
	if cfg.Scheduler == "random" {
		sch = scheduler.NewRandomScheduler(cfg.NumRuns)
	} else {
		sch = scheduler.NewBasicScheduler()
	}

	sm := NewStateManager(
		cfg.GetLocalState,
		cfg.StatesEqual,
	)
	sim := NewSimulator[T, S](sch, sm, cfg.NumRuns, cfg.MaxDepth)
	sender := NewSender(sch)
	sleep := NewSleepManager(sch, sim.NextEvt)
	return SimulationRunner[T, S]{
		sch:          sch,
		sm:           sm,
		sim:          sim,
		SendFactory:  sender.SendFunc,
		SleepFactory: sleep.SleepFunc,
	}
}

type SimulationRunner[T, S any] struct {
	MaxDepth     uint
	SendFactory  func(int) func(int, string, ...any)
	SleepFactory func(int) func(time.Duration)

	InitNodes func() map[int]*T

	StartFuncs     []Request
	IncorrectNodes []int

	Preds []func(GlobalState[S], bool, []GlobalState[S]) bool

	sch scheduler.Scheduler
	sm  *stateManager[T, S]
	sim *Simulator[T, S]
}

func (sr SimulationRunner[T, S]) RunSimulation() {
	// If incorrectNodes is not provided use an empty slice
	if sr.IncorrectNodes == nil {
		sr.IncorrectNodes = make([]int, 0)
	}

	err := sr.sim.Simulate(sr.InitNodes, sr.IncorrectNodes, sr.StartFuncs...)
	if err != nil {
		log.Panicf("Received an error while running simulation: %v", err)
	}

	checker := NewPredicateChecker(sr.Preds...)
	resp := checker.Check(sr.sm.StateRoot)
	_, text := resp.Response()
	fmt.Println(text)

	fmt.Println(sr.sm.StateRoot.Newick())
}
