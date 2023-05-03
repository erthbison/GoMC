package gomc

import (
	"io"
	"log"
	"runtime"

	"gomc/checking"
	"gomc/event"
	"gomc/failureManager"
	"gomc/scheduler"
	"gomc/stateManager"
)

func PrepareSimulation[T, S any](smOpts StateManagerOption[T, S], opts ...SimulatorOption) Simulation[T, S] {
	var (
		maxRuns  = 10000
		maxDepth = 1000
		// number of runs that is simulated at the same time
		numConcurrent = runtime.GOMAXPROCS(0) // Will not change GOMAXPROCS but only return the current value
		// If true will ignore all errors while simulating runs. Will return aggregate of errors at the end. If false will interrupt simulation if an error occur
		ignoreErrors = false
		// If true will ignore panics that occur during the simulation and let them execute as normal, stopping the simulation. If false will catch the panic and return it as an error.
		// ignoring the panic will make it easier to troubleshoot the error since you can use the debugger to inspect the state when it panics. It will also make the simulation stop.
		ignorePanics = false

		sch scheduler.GlobalScheduler
	)

	sch = scheduler.NewPrefix()

	// Use the simulator options to configure
	for _, opt := range opts {
		switch t := opt.(type) {
		case schedulerOption:
			sch = t.sch
		case maxRunsOption:
			maxRuns = t.maxRuns
		case maxDepthOption:
			maxDepth = t.maxDepth
		case numConcurrentOption:
			numConcurrent = t.n
		case ignoreErrorOption:
			ignoreErrors = true
		case ignorePanicOption:
			ignorePanics = true
		}
	}
	sm := smOpts.sm

	sim := NewSimulator(sch, sm, ignoreErrors, ignorePanics, maxRuns, maxDepth, numConcurrent)
	return Simulation[T, S]{
		sim: sim,
		sm:  sm,
	}
}

type Simulation[T, S any] struct {
	sim *Simulator[T, S]
	sm  stateManager.StateManager[T, S]
}

func (sr Simulation[T, S]) Run(InitNodes InitNodeOption[T], requestOpts RequestOption, checker CheckerOption[S], opts ...RunOptions) checking.CheckerResponse {
	// If incorrectNodes is not provided use an empty slice
	var (
		requests = []Request{}

		export []io.Writer

		stopFunc = func(*T) {}

		fm failureManager.FailureManger[T]
	)
	fm = failureManager.NewPerfectFailureManager(func(t *T) {}, []int{})

	for _, opt := range opts {
		switch t := opt.(type) {
		case stopOption[T]:
			stopFunc = t.stop
		case exportOption:
			export = append(export, t.w)
		case failureManagerOption[T]:
			fm = t.fm
		}
	}

	requests = append(requests, requestOpts.request...)
	if len(requests) == 0 {
		log.Panicf("At least one request must be provided to start the simulation")
	}

	// Reset the state and prepare for the simulation
	sr.sm.Reset()

	err := sr.sim.Simulate(fm, InitNodes.f, stopFunc, requests...)
	if err != nil {
		log.Panicf("Received an error while running simulation: %v", err)
	}

	state := sr.sm.State()
	for _, w := range export {
		state.Export(w)
	}

	return checker.checker.Check(state)
}

func PrepareRunner[T, S any](initNodes InitNodeOption[T], getState GetStateOption[T, S], opts ...RunOptions) *Runner[T, S] {
	var (
		stop = func(*T) {}

		eventChanBuffer  = 100
		recordChanBuffer = 100
	)

	for _, opt := range opts {
		switch t := opt.(type) {
		case stopOption[T]:
			stop = t.stop
		case eventChanBufferOption:
			eventChanBuffer = t.size
		case recordChanBufferOption:
			recordChanBuffer = t.size
		}
	}

	r := NewRunner[T, S](
		recordChanBuffer,
	)

	r.Start(
		initNodes.f,
		getState.getState,
		stop,
		eventChanBuffer,
	)
	return r
}

type GetStateOption[T, S any] struct {
	getState func(*T) S
}

func WithStateFunction[T, S any](f func(*T) S) GetStateOption[T, S] {
	return GetStateOption[T, S]{getState: f}
}

type schedulerOption struct {
	sch scheduler.GlobalScheduler
}

func (so schedulerOption) simOpt() {}

// Use a random walk scheduler for the simulation.
//
// The random walk scheduler is a randomized scheduler.
// It uniformly picks the next event to be scheduled from the currently enabled events.
// It does not have a designated stop point, and will continue to schedule events until maxRuns is reached.
// It does not guarantee that all runs have been tested, nor does it guarantee that the same run will not be simulated multiple times.
// Generally, it provides a more even/varied exploration of the state space than systematic exploration
func RandomWalkScheduler(seed int64) SimulatorOption {
	return schedulerOption{sch: scheduler.NewRandom(seed)}
}

// Use a prefix scheduler for the simulation.
//
// The prefix scheduler is a systematic tester, that performs a depth first search of the state space.
// It will stop when the entire state space is explored and will not schedule identical runs.
func PrefixScheduler() SimulatorOption {
	return schedulerOption{sch: scheduler.NewPrefix()}
}

// Use a replay scheduler for the simulation
//
// The replay scheduler replays the provided run, returning an error if it is unable to reproduce it
// The provided run is represented as a slice of event ids, and can be exported using the CheckerResponse.Export()
func ReplayScheduler(run []event.EventId) SimulatorOption {
	return schedulerOption{sch: scheduler.NewReplay(run)}
}

// Use the provided scheduler for the simulation
//
// Used to configure the simulation to use a different implementation of scheduler than is commonly provided
func WithScheduler(sch scheduler.GlobalScheduler) SimulatorOption {
	return schedulerOption{sch: sch}
}

type SimulatorOption interface {
	simOpt()
}

type maxRunsOption struct{ maxRuns int }

func (mro maxRunsOption) simOpt() {}

// Configure the maximum number of runs simulated
//
// Default value is 10000
func MaxRuns(maxRuns int) SimulatorOption {
	return maxRunsOption{maxRuns: maxRuns}
}

type maxDepthOption struct{ maxDepth int }

func (mdo maxDepthOption) simOpt() {}

// Configure the maximum depth explored.
//
// Default value is 1000.
//
// Note that liveness properties can not be verified if a run is not fully explored to its end.
func MaxDepth(maxDepth int) SimulatorOption {
	return maxDepthOption{maxDepth: maxDepth}
}

type numConcurrentOption struct{ n int }

func (nco numConcurrentOption) simOpt() {}

// Configure the number of runs that will be executed concurrently.
//
// Default value is GOMAXPROCS
func NumConcurrent(n int) SimulatorOption {
	return numConcurrentOption{n: n}
}

type ignorePanicOption struct{}

func (ipo ignorePanicOption) simOpt() {}

// Set the ignorePanic flag to true.
//
// If true will ignore panics that occur during the simulation and let them execute as normal, stopping the simulation.
// If false will catch the panic and return it as an error.
// Ignoring the panic will make it easier to troubleshoot the error since you can use the debugger to inspect the state when it panics. It will also make the simulation stop.
func IgnorePanic() SimulatorOption {
	return ignorePanicOption{}
}

type ignoreErrorOption struct{}

func (ieo ignoreErrorOption) simOpt() {}

// Set the ignoreError flag to true.
//
// If true will ignore all errors while simulating runs. Will return aggregate of errors at the end.
// If false will interrupt simulation if an error occur.
func IgnoreError() SimulatorOption {
	return ignoreErrorOption{}
}

type failureManagerOption[T any] struct {
	fm failureManager.FailureManger[T]
}

type RunOptions interface {
	runOpt()
}

func (fmo failureManagerOption[T]) runOpt() {}

func WithFailureManager[T any](fm failureManager.FailureManger[T]) RunOptions {
	return failureManagerOption[T]{fm: fm}
}

func WithPerfectFailureManager[T any](crashFunc func(*T), failingNodes ...int) RunOptions {
	fm := failureManager.NewPerfectFailureManager(
		crashFunc,
		failingNodes,
	)
	return failureManagerOption[T]{fm: fm}
}

type StateManagerOption[T, S any] struct {
	sm stateManager.StateManager[T, S]
}

// Use the provided state manger in the simulation.
func WithStateManager[T, S any](sm stateManager.StateManager[T, S]) StateManagerOption[T, S] {
	return StateManagerOption[T, S]{sm: sm}
}

// Use a TreeStateManager in the simulation.
func WithTreeStateManager[T, S any](getLocalState func(*T) S, statesEqual func(S, S) bool) StateManagerOption[T, S] {
	sm := stateManager.NewTreeStateManager(getLocalState, statesEqual)
	return StateManagerOption[T, S]{sm: sm}
}

// Used to specify how the nodes are started.
type InitNodeOption[T any] struct {
	f func(SimulationParameters) map[int]*T
}

// Uses the provided function f to generate a map of the nodes.
func InitNodeFunc[T any](f func(sp SimulationParameters) map[int]*T) InitNodeOption[T] {
	return InitNodeOption[T]{f: f}
}

// Uses the provided function f to generate nodes with the provided id and add them to a map of the nodes.
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

type CheckerOption[S any] struct {
	checker checking.Checker[S]
}

func WithPredicateChecker[S any](predicates ...checking.Predicate[S]) CheckerOption[S] {
	return CheckerOption[S]{
		checker: checking.NewPredicateChecker(predicates...),
	}
}

func WithChecker[S any](checker checking.Checker[S]) CheckerOption[S] {
	return CheckerOption[S]{checker: checker}
}

type RequestOption struct {
	request []Request
}

func WithRequests(requests ...Request) RequestOption {
	return RequestOption{request: requests}
}

type exportOption struct {
	w io.Writer
}

func (eo exportOption) runOpt() {}

// Write the state to the writer
func Export(w io.Writer) RunOptions {
	return exportOption{w: w}
}

// Configure a function to shut down a node after the execution of a run.
// Default value is an empty function.
type stopOption[T any] struct {
	stop func(*T)
}

func (so stopOption[T]) runOpt() {}

func WithStopFunction[T any](stop func(*T)) RunOptions {
	return stopOption[T]{stop: stop}
}

type recordChanBufferOption struct {
	size int
}

func (opt recordChanBufferOption) runOpt() {}

func RecordChanSize(size int) RunOptions {
	return recordChanBufferOption{size: size}
}

type eventChanBufferOption struct {
	size int
}

func (opt eventChanBufferOption) runOpt() {}

func EventChanBufferSize(size int) RunOptions {
	return eventChanBufferOption{size: size}
}
