package gomc

import (
	"gomc/eventManager"
	"gomc/scheduler"
	"io"
	"log"
	"time"
)

// A struct used to quickly configure the simulation
type SimulationConfig[T, S any] struct {
	// if the value "random" a RandomScheduler is used. Otherwise a BasicScheduler is used
	Scheduler string
	// Number of runs to be used with a RandomScheduler. Default value is 10000. If value is 0 default value is used.
	NumRuns uint
	// Maximum depth of a run. Default value is 1000
	MaxDepth uint
	// The seed of the random scheduler
	Seed int64

	GetLocalState func(*T) S
	StatesEqual   func(S, S) bool
}

func ConfigureSimulation[T, S any](cfg SimulationConfig[T, S]) SimulationRunner[T, S] {
	if cfg.NumRuns == 0 {
		// If numRuns is 0 set to default value 10 000
		cfg.NumRuns = 10000
	}
	if cfg.MaxDepth == 0 {
		cfg.MaxDepth = 1000
	}
	var sch scheduler.Scheduler
	switch cfg.Scheduler {
	case "random":
		sch = scheduler.NewRandomScheduler(cfg.NumRuns, cfg.Seed)
	case "basic":
		sch = scheduler.NewBasicScheduler()
	default:
		sch = scheduler.NewQueueScheduler()
	}

	sm := NewTreeStateManager(
		cfg.GetLocalState,
		cfg.StatesEqual,
	)
	sim := NewSimulator[T, S](sch, sm, cfg.NumRuns, cfg.MaxDepth)
	sender := eventManager.NewSender(sch)
	sleep := eventManager.NewSleepManager(sch, sim.NextEvt)
	return SimulationRunner[T, S]{
		sch: sch,
		sm:  sm,
		sim: sim,

		SendFactory:   sender.SendFunc,
		SleepFactory:  sleep.SleepFunc,
		CrashCallback: sim.Fm.Subscribe,
	}
}

type SimulationRunner[T, S any] struct {
	SendFactory   func(int) func(int, string, ...any)
	SleepFactory  func(int) func(time.Duration)
	CrashCallback func(func(int))

	IncorrectNodes []int

	Preds []Predicate[S]

	sch scheduler.Scheduler
	sm  *treeStateManager[T, S]
	sim *Simulator[T, S]
}

func (sr SimulationRunner[T, S]) RunSimulation(InitNodes func() map[int]*T, StartFuncs ...Request) CheckerResponse {
	// If incorrectNodes is not provided use an empty slice
	if sr.IncorrectNodes == nil {
		sr.IncorrectNodes = make([]int, 0)
	}

	err := sr.sim.Simulate(InitNodes, sr.IncorrectNodes, StartFuncs...)
	if err != nil {
		log.Panicf("Received an error while running simulation: %v", err)
	}

	checker := NewPredicateChecker(sr.Preds...)
	return checker.Check(sr.sm.StateRoot)
}

func (sr SimulationRunner[T, S]) WriteStateTree(wrt io.Writer) {
	sr.sm.Export(wrt)
}
