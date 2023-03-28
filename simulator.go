package gomc

import (
	"errors"
	"fmt"
	"gomc/event"
	"gomc/scheduler"
	"runtime/debug"
)

/*
	Requirements of nodes:
		- Must call the Send function with the required input when sending a message.
		- The message type in the Send function must correspond to a method of the node, whose arguments must match the args passed to the send function.
		- All functions must run to completion without waiting for a response from the tester
*/

type simulationError struct {
	errorSlice []error
}

func (se simulationError) Error() string {
	return fmt.Sprintf("Simulator: %v Errors occurred running simulations. \nError 1: %v", len(se.errorSlice), se.errorSlice[0])
}

type Simulator[T any, S any] struct {

	// The scheduler keeps track of the events and selects the next event to be executed
	Scheduler scheduler.GlobalScheduler

	// If true will ignore all errors while simulating runs. Will return aggregate of errors at the end. If false will interrupt simulation if an error occur
	ignoreErrors bool

	// If true will ignore panics that are raised during the simulation. If false will catch the panic and return it as an error.
	ignorePanics bool

	maxRuns       int
	maxDepth      int
	numConcurrent int
}

func NewSimulator[T any, S any](sch scheduler.GlobalScheduler, ignoreErrors bool, ignorePanics bool, maxRuns int, maxDepth int, numConcurrent int) *Simulator[T, S] {
	// Create a crash manager and make the scheduler subscribe to node crash messages

	return &Simulator[T, S]{
		Scheduler: sch,

		ignoreErrors: ignoreErrors,
		ignorePanics: ignorePanics,

		maxRuns:       maxRuns,
		maxDepth:      maxDepth,
		numConcurrent: numConcurrent,
	}
}

// Run the simulations of the algorithm.
// initNodes: is a function that generates the nodes used and returns them in a map with the id as a key and the node as the value
// funcs: is a variadic arguments of functions that will be scheduled as events by the scheduler. These are used to start the execution of the argument and can represent commands or requests to the service.
// At least one function must be provided for the simulation to start. Otherwise the simulator returns an error.
// Simulate returns nil if the it runs to completion or reaches the max number of runs. It returns an error if it was unable to complete the simulation
func (s Simulator[T, S]) Simulate(sm StateManager[T, S], initNodes func(SimulationParameters) map[int]*T, failingNodes []int, crashFunc func(*T), requests ...Request) error {
	if len(requests) < 1 {
		return fmt.Errorf("Simulator: At least one request should be provided to start simulation.")
	}

	// Pack the parameters into a runParameter to make it easier to handle
	cfg := &runParameters[T]{
		initNodes:    initNodes,
		failingNodes: failingNodes,
		crashFunc:    crashFunc,
		requests:     requests,
	}

	// Used to signal to start the next run
	nextRun := make(chan bool)
	// used by runSimulators to signal that a run has been completed to the main loop. Errors are also returned
	status := make(chan error)
	// Used by the runSimulators to signal that they have stopped executing runs and have closed the goroutine
	// Main loop stops when all runSimulators have stopped executing runs
	closing := make(chan bool)

	ongoing := 0
	startedRuns := 0
	for i := 0; i < s.numConcurrent; i++ {
		ongoing++
		rsim := newRunSimulator(s.Scheduler.GetRunScheduler(), sm.GetRunStateManager(), s.maxDepth, s.ignorePanics)
		go rsim.SimulateRuns(nextRun, status, closing, cfg)

		// Send a signal to start processing runs
		startedRuns++
		nextRun <- true
	}

	return s.mainLoop(ongoing, startedRuns, nextRun, status, closing)
}

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

type runSimulator[T, S any] struct {
	sch scheduler.RunScheduler
	sm  *RunStateManager[T, S]
	fm  *failureManager

	nextEvt chan error

	maxDepth     int
	ignorePanics bool
}

type SimulationParameters struct {
	NextEvt chan error
	Fm      *failureManager
	Sch     scheduler.RunScheduler
}

// Stores the parameters used to start a run.
// Should be read only.
type runParameters[T any] struct {
	initNodes    func(sp SimulationParameters) map[int]*T
	failingNodes []int
	crashFunc    func(*T)
	requests     []Request
}

func newRunSimulator[T, S any](sch scheduler.RunScheduler, sm *RunStateManager[T, S], maxDepth int, ignorePanics bool) *runSimulator[T, S] {
	return &runSimulator[T, S]{
		sch: sch,
		sm:  sm,
		fm:  newFailureManager(),

		nextEvt: make(chan error),

		maxDepth:     maxDepth,
		ignorePanics: ignorePanics,
	}
}

// Main loop of the runSimulator.
// Continuously listens to the nextRun channel and starts simulating a new run each time it receives a signal.
// Stops simulating runs when the channel is closed or when a scheduler.NoRunsError is returned.
// Sends the status of each run on the status channel.
// When it closes it sends an indication on the closing channel
func (rs *runSimulator[T, S]) SimulateRuns(nextRun chan bool, status chan error, closing chan bool, cfg *runParameters[T]) {
	// Continue executing runs until the nextRun channel is closed or until the scheduler returns NoRunsError
	for range nextRun {
		err := rs.simulateRun(cfg)
		if errors.Is(err, scheduler.NoRunsError) {
			break
		}
		// Send error to main loop
		status <- err
	}

	// Indicate that the runSimulator has stopped
	closing <- true
}

func (rs *runSimulator[T, S]) simulateRun(cfg *runParameters[T]) error {
	nodes, err := rs.initRun(cfg.initNodes, cfg.failingNodes, cfg.crashFunc, cfg.requests...)
	if err != nil {
		return err
	}

	err = rs.executeRun(nodes)
	// End the run

	rs.sch.EndRun()
	rs.sm.EndRun()

	if err != nil {
		return fmt.Errorf("Simulator: An error occurred while simulating a run: %v", err)
	}

	return nil
}

func (rs *runSimulator[T, S]) initRun(initNodes func(sp SimulationParameters) map[int]*T, failingNodes []int, crashFunc func(*T), requests ...Request) (map[int]*T, error) {
	nodes := initNodes(SimulationParameters{
		NextEvt: rs.nextEvt,
		Fm:      rs.fm,
		Sch:     rs.sch,
	})
	nodeSlice := []int{}
	for id := range nodes {
		nodeSlice = append(nodeSlice, id)
	}
	rs.fm.Init(nodeSlice)

	rs.sm.UpdateGlobalState(nodes, rs.fm.CorrectNodes(), nil)

	err := rs.sch.StartRun()
	if err != nil {
		return nil, err
	}

	err = rs.scheduleRequests(requests, nodes)
	if err != nil {
		return nil, err
	}

	// Add crash events to simulation.
	for _, id := range failingNodes {
		if _, ok := nodes[id]; !ok {
			continue
		}
		rs.sch.AddEvent(
			event.NewCrashEvent(id, rs.fm.NodeCrash, func() { crashFunc(nodes[id]) }),
		)
	}

	return nodes, nil
}

// Schedules and executes new events until either the scheduler returns a RunEndedError or there is an error during execution of an event.
// If there is an error during the execution it returns the error, otherwise it returns nil
// Uses the state manager to get the global state of the system after the execution of each event
func (rs *runSimulator[T, S]) executeRun(nodes map[int]*T) error {
	depth := 0
	for depth < rs.maxDepth {
		// Select an event
		evt, err := rs.sch.GetEvent()
		if errors.Is(err, scheduler.RunEndedError) {
			return nil
		} else if err != nil {
			return err
		}
		node, ok := nodes[evt.Target()]
		if !ok {
			return fmt.Errorf("Event not targeting an existing node. Targeting %v", evt.Target())
		}
		err = rs.executeEvent(node, evt)
		if err != nil {
			return err
		}
		rs.sm.UpdateGlobalState(nodes, rs.fm.CorrectNodes(), evt)
		depth++
	}
	return nil
}

// Executes the provided event on the provided node in a separate goroutine and returns the error.
// Blocks until the event has been executed and a signal is received on the NextEvt channel
func (rs *runSimulator[T, S]) executeEvent(node *T, evt event.Event) error {
	// execute next event in a goroutine to ensure that we can pause it midway trough if necessary, e.g. for timeouts or some types of messages
	go func() {
		if !rs.ignorePanics {
			// Catch all panics that occur while executing the event. These are often caused by faults in the implementation and are therefore reported to the simulator.
			defer func() {
				if p := recover(); p != nil {
					// using the debug package to get the stack could be useful, but it adds some clutter at the top
					rs.nextEvt <- fmt.Errorf("Node panicked while executing an: %v \nStack Trace:\n %s", p, debug.Stack())
				}
			}()
		}
		evt.Execute(node, rs.nextEvt)
	}()
	return <-rs.nextEvt
}

func (rs *runSimulator[T, S]) scheduleRequests(requests []Request, nodes map[int]*T) error {
	// add all the functions to the scheduler
	addedRequests := 0
	for _, f := range requests {
		if _, ok := nodes[f.Id]; !ok {
			continue
		}
		rs.sch.AddEvent(
			event.NewFunctionEvent(
				addedRequests, f.Id, f.Method, f.Params...,
			),
		)
		addedRequests++
	}
	if addedRequests == 0 {
		return fmt.Errorf("Simulator: At least one request should be provided to start simulation.")
	}
	return nil
}
