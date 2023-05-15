package simulator

import (
	"fmt"
	"gomc/eventManager"
	"gomc/failureManager"
	"gomc/request"
	"gomc/scheduler"
	"gomc/stateManager"
)

// Simulates the a distributed algorithm
//
// Executed all events in the distributed system in a sequential order.
type Simulator[T any, S any] struct {

	// The scheduler keeps track of the events and selects the next event to be executed
	Scheduler scheduler.GlobalScheduler

	sm stateManager.StateManager[T, S]

	// If true will ignore all errors while simulating runs. Will return aggregate of errors at the end. If false will interrupt simulation if an error occur
	ignoreErrors bool

	// If true will ignore panics that are raised during the simulation. If false will catch the panic and return it as an error.
	ignorePanics bool

	maxRuns       int
	maxDepth      int
	numConcurrent int
}

// Create a mew simulator
//
// Configure the simulator with the Scheduler and the StateManager used for the simulation.
//
// ignoreErrors specifies whether to ignore errors. If errors are ignored the simulation will continue simulating runs even if errors occur in some runs.
// A summary of the errors will be provided at the end.
//
// ignorePanics specifies whether to ignore panics. If panics are ignored the simulation will catch panics that occur when executing events and return them.
//
// maxRuns specifies the maximum number of runs to be simulated.
//
// maxDepth specifies the maximum depth of the simulation, i.e. the number of events in a run
//
// numConcurrent specifies the maximum number of runs that are concurrently simulated.
func NewSimulator[T any, S any](sch scheduler.GlobalScheduler, sm stateManager.StateManager[T, S], ignoreErrors bool, ignorePanics bool, maxRuns int, maxDepth int, numConcurrent int) *Simulator[T, S] {
	return &Simulator[T, S]{
		Scheduler: sch,
		sm:        sm,

		ignoreErrors: ignoreErrors,
		ignorePanics: ignorePanics,

		maxRuns:       maxRuns,
		maxDepth:      maxDepth,
		numConcurrent: numConcurrent,
	}
}

// Run the simulations of the algorithm.
//
// fm configures the failure manager used when simulating.
//
// initNodes is a function that generates the nodes used and returns them in a map with the id as a key and the node as the value.
//
// stopFunc is a function specifying how to stop and cleanup the node after the execution of a run.
//
// requests is a variadic arguments of functions that will be scheduled as events by the scheduler. These are used to start the execution of the argument and can represent commands or requests to the service.
// At least one function must be provided for the simulation to start. Otherwise the simulator returns an error.
//
// Simulate returns nil if the it runs to completion or reaches the max number of runs. It returns an error if it was unable to complete the simulation.
func (s Simulator[T, S]) Simulate(fm failureManager.FailureManger[T], initNodes func(eventManager.SimulationParameters) map[int]*T, stopFunc func(*T), requests ...request.Request) error {
	if len(requests) < 1 {
		return fmt.Errorf("Simulator: At least one request should be provided to start simulation.")
	}

	// Pack the parameters into a runParameter to make it easier to handle
	cfg := &runParameters[T]{
		initNodes: initNodes,
		stopNode:  stopFunc,
		requests:  requests,
	}

	// Reset the state of modules so that they are ready for a new simulation
	s.sm.Reset()
	s.Scheduler.Reset()

	// Used to signal to start the next run
	nextRun := make(chan bool)
	// used by runSimulators to signal that a run has been completed to the main loop. Errors are also returned
	status := make(chan error)
	// Used by the runSimulators to signal that they have stopped executing runs and have closed the goroutine
	// Main loop stops when all runSimulators have stopped executing runs
	closing := make(chan bool)

	ongoing := 0
	startedRuns := 0
	for ongoing < s.numConcurrent {
		ongoing++
		rsch := s.Scheduler.GetRunScheduler()
		rsim := newRunSimulator(rsch, s.sm.GetRunStateManager(), fm.GetRunFailureManager(rsch), s.maxDepth, s.ignorePanics)
		go rsim.SimulateRuns(nextRun, status, closing, cfg)

		// Send a signal to start processing runs
		startedRuns++
		nextRun <- true

		if startedRuns >= s.maxRuns {
			break
		}
	}

	return s.mainLoop(ongoing, startedRuns, nextRun, status, closing)
}

// The main loop of the simulation.
//
// Manages the simulation of runs and coordinates the runSimulators.
//
// Receives status updates from each of the runSimulators. One status update for each completed run.
// Processes the status updates and signals for the runSimulator to begin simulating the next run.
// Does not start new simulations if more than maxRuns simulations has been started.
// Returns when all runSimulators has stopped running.
func (s *Simulator[T, S]) mainLoop(ongoing int, startedRuns int, nextRun chan bool, status chan error, closing chan bool) error {
	errorSlice := []error{}
	var out error

	// Stop the simulation by closing the nextRun channel if it is not already closed
	stopped := false
	stop := func() {
		if !stopped {
			stopped = true
			close(nextRun)
		}
	}
	// Loop until all runSimulators has stopped simulating
	for ongoing > 0 {
		select {
		case err := <-status:
			// Handle errors depending on whether the ignoreErrors flag is set or not
			if err != nil {
				if !s.ignoreErrors {
					out = err
					stop()
					break
				} else {
					errorSlice = append(errorSlice, err)
				}
			}

			if startedRuns < s.maxRuns {
				nextRun <- true
				startedRuns++
			} else {
				stop()
			}
		case <-closing:
			ongoing--
		}
	}

	stop()

	// Can safely close the closing and status channels, since we know that all runSimulators has completed and will not try to send on them
	close(closing)
	close(status)

	if s.ignoreErrors && len(errorSlice) > 0 {
		return simulationError{
			errorSlice: errorSlice,
		}
	}
	return out
}
