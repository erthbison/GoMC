package simulator

import (
	"errors"
	"fmt"
	"gomc/event"
	"gomc/eventManager"
	"gomc/failureManager"
	"gomc/request"
	"gomc/scheduler"
	"gomc/stateManager"
	"runtime/debug"
)

// Performs the simulation of runs
type runSimulator[T, S any] struct {
	sch scheduler.RunScheduler
	sm  *stateManager.RunStateManager[T, S]
	fm  failureManager.RunFailureManager[T]

	nextEvt chan error

	maxDepth     int
	ignorePanics bool
}

// create a new runSimulator
//
// Configure a new runSimulator with a runScheduler, runStateManager and a RunFailureManager.
// Also specify the max depth of the run and whether to ignore panics that occur when executing an event.
func newRunSimulator[T, S any](sch scheduler.RunScheduler, sm *stateManager.RunStateManager[T, S], fm failureManager.RunFailureManager[T], maxDepth int, ignorePanics bool) *runSimulator[T, S] {
	return &runSimulator[T, S]{
		sch: sch,
		sm:  sm,
		fm:  fm,

		nextEvt: make(chan error),

		maxDepth:     maxDepth,
		ignorePanics: ignorePanics,
	}
}

// Main loop of the runSimulator.
//
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

// Simulate a run
//
// Simulating consists of three parts: initialization, execution and teardown of the run.
// teardown of the run is always called after the run, even if errors occur.
func (rs *runSimulator[T, S]) simulateRun(cfg *runParameters[T]) error {
	nodes, err := rs.initRun(cfg.initNodes, cfg.requests...)
	if err != nil {
		return fmt.Errorf("Simulator: An error occurred while initializing a run: %w", err)
	}

	// Always teardown the run.
	defer rs.teardownRun(nodes, cfg.stopNode)

	err = rs.executeRun(nodes)
	if err != nil {
		return fmt.Errorf("Simulator: An error occurred while simulating a run: %v", err)
	}
	return nil
}

// Initialization of a run
//
// Creates the nodes and collects the initial state.
// prepares the scheduler and the failure manager for the new run.
// schedules new requests.
func (rs *runSimulator[T, S]) initRun(initNodes func(sp eventManager.SimulationParameters) map[int]*T, requests ...request.Request) (map[int]*T, error) {
	nodes := initNodes(eventManager.SimulationParameters{
		NextEvt:        rs.nextEvent,
		CrashSubscribe: rs.fm.Subscribe,
		EventAdder:     rs.sch,
	})

	rs.sm.UpdateGlobalState(nodes, rs.fm.CorrectNodes(), nil)

	err := rs.sch.StartRun()
	if err != nil {
		return nil, err
	}

	err = rs.scheduleRequests(requests, nodes)
	if err != nil {
		return nil, err
	}

	// Init nodes and schedule crash requests
	rs.fm.Init(nodes)

	return nodes, nil
}

// Teardown the current run.
//
// This includes indicating to the scheduler and state manager that the run has ended.
// Also ensure that all nodes are no longer running.
func (rs *runSimulator[T, S]) teardownRun(nodes map[int]*T, stopFunc func(*T)) {
	// Call end run on scheduler and state manager
	rs.sch.EndRun()
	rs.sm.EndRun()

	// Stop all  nodes
	// What happens if nodes is a nil map???
	// Will not iterate over nil map. No problems.
	for _, node := range nodes {
		stopFunc(node)
	}
}

// Execute the run
//
// Schedules and executes new events until either the scheduler returns a RunEndedError or there is an error during execution of an event.
// If there is any other error during the execution it returns the error, otherwise it returns nil
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

// Execute an event on the provided node
//
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
					rs.nextEvt <- fmt.Errorf("Node panicked while executing an event: %v \nStack Trace:\n %s", p, debug.Stack())
				}
			}()
		}
		evt.Execute(node, rs.nextEvt)
	}()
	return <-rs.nextEvt
}

// Add the requests to the scheduler.
//
// Discards the request if the target node of the request is not a valid node
func (rs *runSimulator[T, S]) scheduleRequests(requests []request.Request, nodes map[int]*T) error {
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
		return fmt.Errorf("At least one request should be provided to start simulation.")
	}
	return nil
}

// Signal the status of the event to the runSimulator.
func (rs *runSimulator[T, S]) nextEvent(status error, _ int) {
	rs.nextEvt <- status
}

// Stores the parameters used to start a run.
// Should be read only.
type runParameters[T any] struct {
	initNodes func(sp eventManager.SimulationParameters) map[int]*T
	stopNode  func(*T)
	requests  []request.Request
}
