package gomc

import (
	"errors"
	"fmt"
	"gomc/event"
	"gomc/scheduler"
)

var (
	NoMessagesError = errors.New("tester: No messages in the queue matching the event")
	UnknownEvent    = errors.New("tester: Received unknown event type")
)

/*
	Requirements of nodes:
		- Must call the Send function with the required input when sending a message.
		- The message type in the Send function must correspond to a method of the node that takes the input (int, int, []byte)
		- All functions must run to completion without waiting for a response from the tester
*/

type Simulator[T any, S any] struct {
	// Responsibility for maintaining the state space.
	sm StateManager[T, S]

	// The event tree should be a tree of events that has been discovered during the traversal of the state space
	// It includes both paths that have been fully explored, and potential paths that we know need further interleaving of the messages to fully explore
	Scheduler scheduler.Scheduler[T]

	// Used to control the flow of events. The simulator will only proceed to gather state and run the next event after it receives a signal on the nextEvt chan
	NextEvt chan error

	fm *failureManager
}

func NewSimulator[T any, S any](sch scheduler.Scheduler[T], sm StateManager[T, S]) *Simulator[T, S] {
	// Create a crash manager and make the scheduler subscribe to node crash messages
	fm := NewFailureManager()
	fm.Subscribe(sch.NodeCrash)
	return &Simulator[T, S]{
		Scheduler: sch,
		sm:        sm,
		fm:        fm,
		NextEvt:   make(chan error),
	}
}

// Run the simulations of the algorithm.
// initNodes: is a function that generates the nodes used and returns them in a map with the id as a key and the node as the value
// funcs: is a variadic arguments of functions that will be scheduled as events by the scheduler. These are used to start the execution of the argument and can represent commands or requests to the service.
// At least one function must be provided for the simulation to start. Otherwise the simulator returns an error.
// Simulate returns nil if the it runs to completion or reaches the max number of runs. It returns an error if it was unable to complete the simulation
func (s Simulator[T, S]) Simulate(initNodes func() map[int]*T, funcs map[int][]func(*T) error, failingNodes []int) error {
	numRuns := 0
	if len(funcs) < 1 {
		return fmt.Errorf("Simulator: Need at least one provided function to start simulation. No functions provided.")
	}
	for numRuns < 50000 {
		// Perform initialization of the run
		nodes := initNodes()
		s.sm.UpdateGlobalState(nodes, s.fm.CorrectNodes())
		nodeSlice := []int{}
		for id := range nodes {
			nodeSlice = append(nodeSlice, id)
		}
		s.fm.InitNodes(nodeSlice)

		// Add all the function events to the scheduler
		num := 0
		for target, funcSlice := range funcs {
			for _, f := range funcSlice {
				s.Scheduler.AddEvent(
					event.NewFunctionEvent(
						num, target, f,
					),
				)
				num++
			}
		}
		if num == 0 {
			return fmt.Errorf("Simulator: Need at least one provided function to start simulation. No functions provided.")
		}

		// Add crash events to simulation.
		for _, id := range failingNodes {
			s.Scheduler.AddEvent(
				event.NewCrashEvent[T](id, s.fm.NodeCrash),
			)
		}

		err := s.executeRun(nodes, s.fm.CorrectNodes())
		if errors.Is(err, scheduler.NoEventError) {
			// If there are no available events that means that all possible event chains have been attempted and we are done
			return nil
		} else if err != nil {
			return fmt.Errorf("Simulator: An error occurred while simulating a run: %v", err)
		}
		// End the run
		// Add an end event at the end of this path of the event tree
		s.Scheduler.EndRun()
		s.sm.EndRun()
		numRuns++
		if numRuns%1000 == 0 {
			fmt.Println("Running Simulation:", numRuns)
		}
	}
	return nil
}

// Schedules and executes new events until either the scheduler returns a RunEndedError or there is an error during execution of an event.
// If there is an error during the execution it returns the error, otherwise it returns nil
// Uses the state manager to get the global state of the system after the execution of each event
func (s *Simulator[T, S]) executeRun(nodes map[int]*T, correct map[int]bool) error {
	for {
		// Select an event
		evt, err := s.Scheduler.GetEvent()
		if errors.Is(err, scheduler.RunEndedError) {
			return nil
		} else if err != nil {
			return err
		}
		if node, ok := nodes[evt.Target()]; !ok {
			return fmt.Errorf("Event not targeting an existing node")
		} else {
			// execute next event in a goroutine to ensure that we can pause it midway trough if necessary, e.g. for timeouts or some types of messages
			go func() {
				// Catch all panics that occur while executing the event. These are often caused by faults in the implementation and are therefore reported to the simulator.
				defer func() {
					if p := recover(); p != nil {
						s.NextEvt <- fmt.Errorf("Error while executing Event: %v", p)
					}
				}()
				evt.Execute(node, s.NextEvt)
			}()
		}
		err = <-s.NextEvt
		if err != nil {
			return err
		}
		s.sm.UpdateGlobalState(nodes, correct)
	}
}
