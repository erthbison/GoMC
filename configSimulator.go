package gomc

import (
	"io"
	"log"
	"runtime"

	"gomc/checking"
	"gomc/config"
	"gomc/event"
	"gomc/eventManager"
	"gomc/failureManager"
	"gomc/request"
	"gomc/scheduler"
	"gomc/simulator"
	"gomc/stateManager"
)

// Prepare simulation with initial configuration.
//
// Initializes the simulator with the necessary parameters.
// See the SimulatorOptions for a full overview of possible options.
// Default values will be used if no value is provided.
// Default scheduler is PrefixScheduler.
func PrepareSimulation[T, S any](smOpts StateManagerOption[T, S], opts ...SimulatorOption) Simulation[T, S] {
	var (
		// Maximum number of runs simulated
		maxRuns = 10000

		// Maximum number of events in a run
		maxDepth = 100

		// number of runs that is simulated at the same time
		numConcurrent = runtime.GOMAXPROCS(0) // Will not change GOMAXPROCS but only return the current value

		// If true will ignore all errors while simulating runs. Will return aggregate of errors at the end. If false will interrupt simulation if an error occur
		ignoreErrors = false

		// If true will ignore panics that occur during the simulation and let them execute as normal, stopping the simulation. If false will catch the panic and return it as an error.
		// ignoring the panic will make it easier to troubleshoot the error since you can use the debugger to inspect the state when it panics. It will also make the simulation stop.
		ignorePanics = false

		sch scheduler.GlobalScheduler
	)

	// Use the simulator options to configure
	for _, opt := range opts {
		switch t := opt.(type) {
		case config.SchedulerOption:
			sch = t.Sch
		case config.MaxRunsOption:
			maxRuns = t.MaxRuns
		case config.MaxDepthOption:
			maxDepth = t.MaxDepth
		case config.NumConcurrentOption:
			numConcurrent = t.N
		case config.IgnoreErrorOption:
			ignoreErrors = true
		case config.IgnorePanicOption:
			ignorePanics = true
		}
	}
	if sch == nil {
		sch = scheduler.NewPrefix()
	}

	sm := smOpts.sm

	sim := simulator.NewSimulator(sch, sm, ignoreErrors, ignorePanics, maxRuns, maxDepth, numConcurrent)
	return Simulation[T, S]{
		sim: sim,
		sm:  sm,
	}
}

// Stores the configured Simulator.
//
// Can be used to run multiple simulations.
// A simulation is started by calling the Run method.
// Only one simulation can be run at a time.
type Simulation[T, S any] struct {
	sim *simulator.Simulator[T, S]
	sm  stateManager.StateManager[T, S]
}

// Run the simulation of the algorithm.
//
// The InitNodeOption, requestOption and CheckerOptions are mandatory.
// All RunOptions are optional. Default values will be used if no values are provided.
//
// Returns a checking.CheckerResponse type containing the results of the simulation
func (sr Simulation[T, S]) Run(InitNodes InitNodeOption[T], requestOpts RequestOption, checker CheckerOption[S], opts ...RunOptions) checking.CheckerResponse {
	// If incorrectNodes is not provided use an empty slice
	var (
		requests = []request.Request{}

		export []io.Writer

		stopFunc = func(*T) {}

		fm failureManager.FailureManger[T]
	)

	for _, opt := range opts {
		switch t := opt.(type) {
		case config.StopOption[T]:
			stopFunc = t.Stop
		case config.ExportOption:
			export = append(export, t.W)
		case config.FailureManagerOption[T]:
			fm = t.Fm
		}
	}

	if fm == nil {
		fm = failureManager.NewPerfectFailureManager(func(t *T) {}, []int{})
	}

	requests = append(requests, requestOpts.request...)
	if len(requests) == 0 {
		log.Panicf("At least one request must be provided to start the simulation")
	}

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

// A option used to configure the Simulator
type SimulatorOption interface {
	// noop method
	SimOpt()
}

// Use a random walk scheduler for the simulation.
//
// The random walk scheduler is a randomized scheduler.
// It uniformly picks the next event to be scheduled from the currently enabled events.
// It does not have a designated stop point, and will continue to schedule events until maxRuns is reached.
// It does not guarantee that all runs have been tested, nor does it guarantee that the same run will not be simulated multiple times.
// Generally, it provides a more even/varied exploration of the state space than systematic exploration
func RandomWalkScheduler(seed int64) SimulatorOption {
	return config.SchedulerOption{Sch: scheduler.NewRandom(seed)}
}

// Use a prefix scheduler for the simulation.
//
// The prefix scheduler is a systematic tester, that performs a depth first search of the state space.
// It will stop when the entire state space is explored and will not schedule identical runs.
func PrefixScheduler() SimulatorOption {
	return config.SchedulerOption{Sch: scheduler.NewPrefix()}
}

// Use a replay scheduler for the simulation
//
// The replay scheduler replays the provided run, returning an error if it is unable to reproduce it
// The provided run is represented as a slice of event ids, and can be exported using the CheckerResponse.Export()
func ReplayScheduler(run []event.EventId) SimulatorOption {
	return config.SchedulerOption{Sch: scheduler.NewReplay(run)}
}

// Use the provided scheduler for the simulation
//
// Used to configure the simulation to use a different implementation of scheduler than is commonly provided
func WithScheduler(sch scheduler.GlobalScheduler) SimulatorOption {
	return config.SchedulerOption{Sch: sch}
}

// Configure the maximum number of runs simulated
//
// Default value is 10000
func MaxRuns(maxRuns int) SimulatorOption {
	return config.MaxRunsOption{MaxRuns: maxRuns}
}

// Configure the maximum depth explored.
//
// Default value is 100.
//
// Note that liveness properties can not be verified if a run is not fully explored to its end.
func MaxDepth(maxDepth int) SimulatorOption {
	return config.MaxDepthOption{MaxDepth: maxDepth}
}

// Configure the number of runs that will be executed concurrently.
//
// Default value is GOMAXPROCS
func NumConcurrent(n int) SimulatorOption {
	return config.NumConcurrentOption{N: n}
}

// Set the ignorePanic flag to true.
//
// If true will ignore panics that occur during the simulation and let them execute as normal, stopping the simulation.
// If false will catch the panic and return it as an error.
// Ignoring the panic will make it easier to troubleshoot the error since you can use the debugger to inspect the state when it panics. It will also make the simulation stop.
func IgnorePanic() SimulatorOption {
	return config.IgnorePanicOption{}
}

// Set the ignoreError flag to true.
//
// If true will ignore all errors while simulating runs. Will return aggregate of errors at the end.
// If false will interrupt simulation if an error occur.
func IgnoreError() SimulatorOption {
	return config.IgnoreErrorOption{}
}

// Optional parameters used to configure a simulation
type RunOptions interface {
	RunOpt()
}

// Specify the failure manager used for the Simulation
//
// Default value is a PerfectFailureManager with no node crashes.
func WithFailureManager[T any](fm failureManager.FailureManger[T]) RunOptions {
	return config.FailureManagerOption[T]{Fm: fm}
}

// Configure the simulation to use a PerfectFailureManager.
//
// The PerfectFailureManager implements the crash-stop failures in a synchronous system.
// It imitates the behavior of the Perfect Failure Detector.
//
// Default value is a PerfectFailureManager with no node crashes.
func WithPerfectFailureManager[T any](crashFunc func(*T), failingNodes ...int) RunOptions {
	fm := failureManager.NewPerfectFailureManager(
		crashFunc,
		failingNodes,
	)
	return config.FailureManagerOption[T]{Fm: fm}
}

// Configure the StateManager used to manage the state of the distributed system.
//
// The State Manager collects and manages the state of the system under testing.
type StateManagerOption[T, S any] struct {
	sm stateManager.StateManager[T, S]
}

// Use the provided state manger in the simulation.
func WithStateManager[T, S any](sm stateManager.StateManager[T, S]) StateManagerOption[T, S] {
	return StateManagerOption[T, S]{sm: sm}
}

// Use a TreeStateManager in the simulation.
//
// The TreeStateManager organizes the state in a tree structure, which is stored in memory.
// The TreeStateManager is configured with a function collecting the local state from a node and a function checking the equality of two local states.
func WithTreeStateManager[T, S any](getLocalState func(*T) S, statesEqual func(S, S) bool) StateManagerOption[T, S] {
	sm := stateManager.NewTreeStateManager(getLocalState, statesEqual)
	return StateManagerOption[T, S]{sm: sm}
}

// Configures how the nodes are started.
//
// The function should create the nodes that will be used when running the simulation.
// It should also initialize the Event Manager that will be used.
// The provided SimulationParameters should be used to configure the Event Managers with run specific data.
type InitNodeOption[T any] struct {
	f func(eventManager.SimulationParameters) map[int]*T
}

// Uses the provided function f to generate a map of the nodes.
func InitNodeFunc[T any](f func(sp eventManager.SimulationParameters) map[int]*T) InitNodeOption[T] {
	return InitNodeOption[T]{f: f}
}

// Uses the provided function f to generate individual nodes with the provided id and add them to a map of the nodes.
func InitSingleNode[T any](nodeIds []int, f func(id int, sp eventManager.SimulationParameters) *T) InitNodeOption[T] {
	t := func(sp eventManager.SimulationParameters) map[int]*T {
		nodes := map[int]*T{}
		for _, id := range nodeIds {
			nodes[id] = f(id, sp)
		}
		return nodes
	}
	return InitNodeOption[T]{f: t}
}

// Configures the Checker to be used when verifying the algorithm.
//
// The Checker verifies that the properties of the algorithm holds.
// It returns a CheckerResponse with the result of the simulation.
type CheckerOption[S any] struct {
	checker checking.Checker[S]
}

// Use a PredicateChecker to verify the algorithm.
//
// The predicate checker uses functions to define the properties of the algorithm.
// functions are provided as the checking.Predicate type.
func WithPredicateChecker[S any](predicates ...checking.Predicate[S]) CheckerOption[S] {
	return CheckerOption[S]{
		checker: checking.NewPredicateChecker(predicates...),
	}
}

// Specify the Checker used to verify the algorithm.
func WithChecker[S any](checker checking.Checker[S]) CheckerOption[S] {
	return CheckerOption[S]{checker: checker}
}

// Configures the requests to the distributed system.
//
// The request are used to start the simulation and define the scenario of the simulation.
type RequestOption struct {
	request []request.Request
}

// Configures the requests to the distributed system.
//
// The request are used to start the simulation and define the scenario of the simulation.
func WithRequests(requests ...request.Request) RequestOption {
	return RequestOption{request: requests}
}

// Add a writer that the state will be exported to
//
// Can be called multiple times.
// Default value is no writers
func Export(w io.Writer) RunOptions {
	return config.ExportOption{W: w}
}

// Configures a function used to stop the nodes after a run.
//
// The function should clean up all operations of the nodes to avoid memory leaks across runs.
//
// Default value is empty function.
func WithStopFunctionSimulator[T any](stop func(*T)) RunOptions {
	return config.StopOption[T]{Stop: stop}
}
