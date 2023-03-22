package gomc

import (
	"errors"
	"fmt"
	"gomc/event"
	"gomc/scheduler"
)

/*
	Requirements of nodes:
		- Must call the Send function with the required input when sending a message.
		- The message type in the Send function must correspond to a method of the node, whose arguments must match the args passed to the send function.
		- All functions must run to completion without waiting for a response from the tester
*/

type Simulator[T any, S any] struct {

	// The scheduler keeps track of the events and selects the next event to be executed
	Scheduler scheduler.GlobalScheduler

	maxRuns       int
	maxDepth      int
	numConcurrent int
}

func NewSimulator[T any, S any](sch scheduler.GlobalScheduler, maxRuns int, maxDepth int, numConcurrent int) *Simulator[T, S] {
	// Create a crash manager and make the scheduler subscribe to node crash messages

	return &Simulator[T, S]{
		Scheduler: sch,

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
func (s Simulator[T, S]) Simulate(sm StateManager[T, S], initNodes func(sch scheduler.RunScheduler, nextEvt chan error, crashCallback func(func(int))) map[int]*T, failingNodes []int, crashFunc func(*T), requests ...Request) error {
	if len(requests) < 1 {
		return fmt.Errorf("Simulator: At least one request should be provided to start simulation.")
	}

	var numRuns int
	nextRun := make(chan bool)
	runStatus := make(chan error)
	for i := 0; i < s.numConcurrent; i++ {
		go func() {
			rsim := newRunSimulator(s.Scheduler, sm, s.maxDepth)
			for {
				if _, ok := <-nextRun; !ok {
					return
				}
				runStatus <- rsim.SimulateRun(initNodes, failingNodes, crashFunc, requests...)

			}
		}()
		nextRun <- true
	}
	for status := range runStatus {
		if status != nil {
			if errors.Is(status, scheduler.NoEventError) {
				close(nextRun)
				return nil
			} else {
				return status
			}
		}
		numRuns++
		if numRuns > s.maxRuns {
			close(nextRun)
			return nil
		}
		nextRun <- true
	}
	return nil
}

type runSimulator[T, S any] struct {
	sch scheduler.RunScheduler
	sm  *RunStateManager[T, S]
	fm  *failureManager

	NextEvt chan error

	maxDepth int
}

func newRunSimulator[T, S any](sch scheduler.GlobalScheduler, sm StateManager[T, S], maxDepth int) *runSimulator[T, S] {
	return &runSimulator[T, S]{
		sch: sch.GetRunScheduler(),
		sm:  sm.GetRunStateManager(),
		fm:  NewFailureManager(),

		NextEvt: make(chan error),

		maxDepth: maxDepth,
	}
}

func (rs *runSimulator[T, S]) SimulateRun(initNodes func(sch scheduler.RunScheduler, nextEvt chan error, crashCallback func(func(int))) map[int]*T, failingNodes []int, crashFunc func(*T), requests ...Request) error {
	nodes := initNodes(rs.sch, rs.NextEvt, rs.fm.Subscribe)
	nodeSlice := []int{}
	for id := range nodes {
		nodeSlice = append(nodeSlice, id)
	}
	rs.fm.Init(nodeSlice)

	rs.sm.UpdateGlobalState(nodes, rs.fm.CorrectNodes(), nil)

	err := rs.sch.StartRun()
	if err != nil {
		return err
	}

	rs.scheduleRequests(requests)

	// Add crash events to simulation.
	for _, id := range failingNodes {
		rs.sch.AddEvent(
			event.NewCrashEvent(id, rs.fm.NodeCrash, func() { crashFunc(nodes[id]) }),
		)
	}

	err = rs.executeRun(nodes)
	if err != nil {
		return fmt.Errorf("Simulator: An error occurred while simulating a run: %v", err)
	}
	// End the run
	// Add an end event at the end of this path of the event tree
	rs.sch.EndRun()
	rs.sm.EndRun()
	return nil
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
		// Catch all panics that occur while executing the event. These are often caused by faults in the implementation and are therefore reported to the simulator.
		// defer func() {
		// 	if p := recover(); p != nil {
		// 		// using the debug package to get the stack could be useful, but it adds some clutter at the top
		// 		s.NextEvt <- fmt.Errorf("Node panicked while executing event: %v \nStack Trace:\n %s", p, debug.Stack())
		// 	}
		// }()
		evt.Execute(node, rs.NextEvt)
	}()
	return <-rs.NextEvt
}

func (rs *runSimulator[T, S]) scheduleRequests(requests []Request) {
	// add all the functions to the scheduler
	for i, f := range requests {
		rs.sch.AddEvent(
			event.NewFunctionEvent(
				i, f.Id, f.Method, f.Params...,
			),
		)
	}
}
