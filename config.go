package gomc

import (
	"gomc/scheduler"
	"io"
	"log"
	"runtime"
)

func Prepare[T, S any](schOpt SchedulerOption, opts ...SimulatorOption) SimulationRunner[T, S] {
	// Default values:
	var (
		maxRuns       = 1000
		maxDepth      = 1000
		numConcurrent = runtime.GOMAXPROCS(0) // Will not change GOMAXPROCS but only return the current value
	)

	// Use the simulator options to configure
	for _, opt := range opts {
		switch t := opt.(type) {
		case MaxRunsOption:
			maxRuns = t.maxRuns
		case MaxDepthOption:
			maxDepth = t.maxDepth
		case NumConcurrentOption:
			numConcurrent = t.n
		}
	}
	sch := schOpt.sch
	sim := NewSimulator[T, S](sch, maxRuns, maxDepth, numConcurrent)
	return SimulationRunner[T, S]{
		sch: sch,
		sim: sim,
	}
}

type SimulationRunner[T, S any] struct {
	sch scheduler.GlobalScheduler
	sim *Simulator[T, S]
}

func (sr SimulationRunner[T, S]) RunSimulation(InitNodes InitNodeOption[T], requestOpts RequestOption, smOpts StateManagerOption[T, S], opts ...RunOptions) CheckerResponse {
	// If incorrectNodes is not provided use an empty slice
	var (
		incorrectNodes = []int{}
		predicates     = []Predicate[S]{}
		requests       = []Request{}
		crashFunc      = func(*T) {}

		export io.Writer
	)

	for _, opt := range opts {
		switch t := opt.(type) {
		case IncorrectNodesOption[T]:
			incorrectNodes = append(incorrectNodes, t.nodes...)
			crashFunc = t.crashFunc
		case PredicateOption[S]:
			predicates = append(predicates, t.pred...)
		case ExportOption:
			export = t.w
		}
	}

	requests = append(requests, requestOpts.request...)
	if len(requests) == 0 {
		log.Panicf("At least one request must be provided to start the simulation")
	}

	sm := smOpts.sm
	err := sr.sim.Simulate(sm, InitNodes.f, incorrectNodes, crashFunc, requests...)
	if err != nil {
		log.Panicf("Received an error while running simulation: %v", err)
	}

	if export != nil {
		sm.State().Export(export)
	}

	// Check the predicates
	checker := NewPredicateChecker(predicates...)
	return checker.Check(sm.State())
}

type SchedulerOption struct {
	sch scheduler.GlobalScheduler
}

func RandomWalkScheduler(seed int64) SchedulerOption {
	return SchedulerOption{sch: scheduler.NewRandom(seed)}
}

func PrefixScheduler() SchedulerOption {
	return SchedulerOption{sch: scheduler.NewPrefix()}
}

// func BasicScheduler() SchedulerOption {
// 	return SchedulerOption{sch: scheduler.NewBasicScheduler()}
// }

func ReplayScheduler(run []string) SchedulerOption {
	return SchedulerOption{sch: scheduler.NewReplay(run)}
}

func WithScheduler(sch scheduler.GlobalScheduler) SchedulerOption {
	return SchedulerOption{sch: sch}
}

type SimulatorOption interface{}

type MaxRunsOption struct{ maxRuns int }

func MaxRuns(maxRuns int) SimulatorOption {
	return MaxRunsOption{maxRuns: maxRuns}
}

type MaxDepthOption struct{ maxDepth int }

func MaxDepth(maxDepth int) SimulatorOption {
	return MaxDepthOption{maxDepth: maxDepth}
}

type NumConcurrentOption struct{ n int }

func NumConcurrent(n int) SimulatorOption {
	return NumConcurrentOption{n: n}
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
	f func(SimulationParameters) map[int]*T
}

// Uses the provided function f to generate a map of the nodes
func InitNodeFunc[T any](f func(sp SimulationParameters) map[int]*T) InitNodeOption[T] {
	return InitNodeOption[T]{f: f}
}

// Uses the provided function f to generate nodes with the provided id and add them to a map of the nodes
func InitSingleNode[T any](nodeIds []int, f func(id int, sp SimulationParameters) *T) InitNodeOption[T] {
	t := func(sp SimulationParameters) map[int]*T {
		nodes := map[int]*T{}
		for _, id := range nodeIds {
			nodes[id] = f(id, sp)
		}
		return nodes
	}
	return InitNodeOption[T]{f: t}
}

type RunOptions interface{}

type IncorrectNodesOption[T any] struct {
	crashFunc func(*T)
	nodes     []int
}

func IncorrectNodes[T any](crashFunc func(*T), nodes ...int) RunOptions {
	return IncorrectNodesOption[T]{crashFunc: crashFunc, nodes: nodes}
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

type ExportOption struct {
	w io.Writer
}

func Export(w io.Writer) RunOptions {
	return ExportOption{w: w}
}
