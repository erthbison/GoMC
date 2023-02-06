package gomc

import (
	"errors"
	"fmt"
	"runtime"
	"time"
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
	Scheduler Scheduler[T]

	nextEvt      chan error
	timeoutChans map[string]chan time.Time

	end     bool
	numRuns int
}

func NewSimulator[T any, S any](sch Scheduler[T], sm StateManager[T, S]) *Simulator[T, S] {
	return &Simulator[T, S]{
		Scheduler:    sch,
		sm:           sm,
		nextEvt:      make(chan error),
		numRuns:      0,
		end:          false,
		timeoutChans: make(map[string]chan time.Time),
	}
}

func (s Simulator[T, S]) Simulate(initNodes func() map[int]*T, funcs ...func(map[int]*T) error) error {
	for s.numRuns < 50000 {
		// Perform initialization of the run
		nodes := initNodes()
		s.sm.UpdateGlobalState(nodes)

		// Add all the function events to the scheduler
		for i, f := range funcs {
			s.Scheduler.AddEvent(
				FunctionEvent[T]{
					index: i,
					F:     f,
				},
			)
		}

		err := s.executeRun(nodes)
		if errors.Is(err, NoEventError) {
			// If there are no available events that means that all possible event chains have been attempted and we are done
			return nil
		} else if err != nil {
			return fmt.Errorf("An error occurred while scheduling the next event: %w", err)
		}
		// End the run
		// Add an end event at the end of this path of the event tree
		s.Scheduler.EndRun()
		s.sm.EndRun()
		s.numRuns++
		if s.numRuns%1000 == 0 {
			fmt.Println("Running Simulation:", s.numRuns)
		}
	}
	return nil
}

func (s *Simulator[T, S]) executeRun(nodes map[int]*T) error {
	for {
		// Select an event
		evt, err := s.Scheduler.GetEvent()
		if errors.Is(err, RunEndedError) {
			return nil
		} else if err != nil {
			return err
		}
		// execute next event in a goroutine to ensure that we can pause it midway trough if necessary, e.g. for timeouts or some types of messages
		go evt.Execute(nodes, s.nextEvt)
		err = <-s.nextEvt
		if err != nil {
			return fmt.Errorf("An error occurred while executing the next event: %w", err)
		}
		s.sm.UpdateGlobalState(nodes)
	}
}

func (t *Simulator[T, S]) Send(from, to int, msgType string, msg []byte) {
	t.Scheduler.AddEvent(MessageEvent[T]{
		From:  from,
		To:    to,
		Type:  msgType,
		Value: msg,
	})
}

func (t *Simulator[T, S]) Timeout(_ time.Duration) <-chan time.Time {
	_, file, line, _ := runtime.Caller(1)
	evt := NewTimeoutEvent[T](fmt.Sprintf("File: %v, Line: %v", file, line), t.timeoutChans)
	t.Scheduler.AddEvent(evt)
	// Inform the simulator that the process is currently waiting for a scheduled timeout
	// The simulator can now proceed with scheduling events
	t.nextEvt <- nil
	return t.timeoutChans[evt.Id()]
}
