package gomc

import (
	"gomc/eventManager"
	"gomc/scheduler"
	"log"
	"time"
)

func Prepare[T, S any](schOpt SchedulerOption, opts ...SimulatorOption) SimulationRunner[T, S] {
	// Default values:
	var (
		maxRuns  = uint(1000)
		maxDepth = uint(1000)
	)

	// Use the simulator options to configure
	for _, opt := range opts {
		switch t := opt.(type) {
		case MaxRunsOption:
			maxRuns = t.maxRuns
		case MaxDepthOption:
			maxDepth = t.maxDepth
		}
	}
	sch := schOpt.sch

	sim := NewSimulator[T, S](sch, maxRuns, maxDepth)
	sender := eventManager.NewSender(sch)
	sleep := eventManager.NewSleepManager(sch, sim.NextEvt)
	return SimulationRunner[T, S]{
		sch: sch,
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

	sch scheduler.Scheduler
	sim *Simulator[T, S]
}

func (sr SimulationRunner[T, S]) RunSimulation(InitNodes InitNodeOption[T], requestOpts RequestOption, smOpts StateManagerOption[T, S], opts ...RunOptions) CheckerResponse {
	// If incorrectNodes is not provided use an empty slice
	var (
		incorrectNodes = []int{}
		predicates     = []Predicate[S]{}
		requests       = []Request{}
	)

	for _, opt := range opts {
		switch t := opt.(type) {
		case IncorrectNodesOption:
			incorrectNodes = append(incorrectNodes, t.nodes...)
		case PredicateOption[S]:
			predicates = append(predicates, t.pred...)
		}
	}

	requests = append(requests, requestOpts.request...)
	if len(requests) == 0 {
		log.Panicf("At least one request must be provided to start the simulation")
	}

	sm := smOpts.sm
	err := sr.sim.Simulate(sm, InitNodes.f, incorrectNodes, requests...)
	if err != nil {
		log.Panicf("Received an error while running simulation: %v", err)
	}
	// Check the predicates
	checker := NewPredicateChecker(predicates...)
	return checker.Check(sm.State())
}

func (sr *SimulationRunner[T, S]) Scheduler() scheduler.Scheduler {
	return sr.sch
}

func (sr *SimulationRunner[T, S]) NextEvent() chan error {
	return sr.sim.NextEvt
}

type SchedulerOption struct {
	sch scheduler.Scheduler
}

func RandomWalkScheduler(maxRuns uint, seed int64) SchedulerOption {
	return SchedulerOption{sch: scheduler.NewRandomScheduler(maxRuns, seed)}
}

func QueueScheduler() SchedulerOption {
	return SchedulerOption{sch: scheduler.NewQueueScheduler()}
}

func BasicScheduler() SchedulerOption {
	return SchedulerOption{sch: scheduler.NewBasicScheduler()}
}

func ReplayScheduler(run []string) SchedulerOption {
	return SchedulerOption{sch: scheduler.NewReplayScheduler(run)}
}

func WithScheduler(sch scheduler.Scheduler) SchedulerOption {
	return SchedulerOption{sch: sch}
}

type SimulatorOption interface{}

type MaxRunsOption struct{ maxRuns uint }

func MaxRuns(maxRuns uint) SimulatorOption {
	return MaxRunsOption{maxRuns: maxRuns}
}

type MaxDepthOption struct{ maxDepth uint }

func MaxDepth(maxDepth uint) SimulatorOption {
	return MaxDepthOption{maxDepth: maxDepth}
}

type StateManagerOption[T, S any] struct{ sm StateManager[T, S] }

func WithStateManager[T, S any](sm StateManager[T, S]) StateManagerOption[T, S] {
	return StateManagerOption[T, S]{sm: sm}
}

func WithTreeStateManager[T, S any](getLocalState func(*T) S, statesEqual func(S, S) bool) StateManagerOption[T, S] {
	sm := NewTreeStateManager(getLocalState, statesEqual)
	return StateManagerOption[T, S]{sm: sm}
}

// Used to specify how the nodes are started.
type InitNodeOption[T any] struct {
	f func() map[int]*T
}

// Uses the provided function f to generate a map of the nodes
func InitNodeFunc[T any](f func() map[int]*T) InitNodeOption[T] {
	return InitNodeOption[T]{f: f}
}

// Uses the provided function f to generate nodes with the provided id and add them to a map of the nodes
func InitSingleNode[T any](nodeIds []int, f func(id int) *T) InitNodeOption[T] {
	t := func() map[int]*T {
		nodes := map[int]*T{}
		for _, id := range nodeIds {
			nodes[id] = f(id)
		}
		return nodes
	}
	return InitNodeOption[T]{f: t}
}

type RunOptions interface{}

type IncorrectNodesOption struct{ nodes []int }

func IncorrectNodes(nodes ...int) RunOptions {
	return IncorrectNodesOption{nodes: nodes}
}

type PredicateOption[S any] struct{ pred []Predicate[S] }

func WithPredicate[S any](preds ...Predicate[S]) RunOptions {
	return PredicateOption[S]{pred: preds}
}

type RequestOption struct {
	request []Request
}

func WithRequests(requests ...Request) RequestOption {
	return RequestOption{request: requests}
}
