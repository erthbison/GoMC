package tester

import (
	"errors"
	"fmt"
	"log"
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
	nodes map[int]*T

	// Responsibility for maintaining the state space.
	sm StateManager[T, S]

	// The event tree should be a tree of events that has been discovered during the traversal of the state space
	// It includes both paths that have been fully explored, and potential paths that we know need further interleaving of the messages to fully explore
	Scheduler Scheduler[T]

	end     bool
	numRuns int
}

func NewSimulator[T any, S any](sch Scheduler[T], sm StateManager[T, S]) *Simulator[T, S] {
	return &Simulator[T, S]{
		nodes:     map[int]*T{},
		Scheduler: sch,
		sm:        sm,
		numRuns:   0,
		end:       false,
	}
}

func (t *Simulator[T, S]) Simulate(initNodes func() map[int]*T, funcs ...func(map[int]*T) error) {
	// t.getLocalState = getLocalState
outer:
	for !t.isCompleted() {
		// Create nodes and init states for this run
		t.nodes = initNodes()
		t.sm.UpdateGlobalState(t.nodes)

		// Add all the function events to the scheduler
		for i, f := range funcs {
			t.Scheduler.AddEvent(
				FunctionEvent[T]{
					index: i,
					F:     f,
				},
			)
		}
		for {
			err := t.executeNextEvent()
			if err != nil {
				if errors.Is(err, NoEventError) {
					// If there are no available events that means that all possible event chains have been attempted and we are done
					// Update the end flag
					t.end = true
					break outer
				} else if errors.Is(err, RunEndedError) {
					break
				} else {
					log.Panicf("An error occurred while scheduling the next message: %v", err)
				}
				break
			}
			t.sm.UpdateGlobalState(t.nodes)
		}
		// Add an end event at the end of this path of the event tree
		t.Scheduler.EndRun()
		t.sm.EndRun()

		t.numRuns++
		if t.numRuns%1000 == 0 {
			fmt.Println("Running Simulation:", t.numRuns)
		}
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

func (t *Simulator[T, S]) executeNextEvent() error {
	event, err := t.Scheduler.GetEvent()
	if err != nil {
		return err
	}
	return event.Execute(t.nodes)
	// switch evt := event.(type) {
	// case MessageEvent:
	// 	t.sendMessage(evt)
	// case FunctionEvent[T]:
	// 	evt.F(t.nodes)
	// case TimeoutEvent:
	// 	evt.Timeout <- time.Time{}
	// default:
	// 	return UnknownEvent
	// }
	// return nil
}

// func (t *Simulator[T, S]) sendMessage(evt MessageEvent) {
// 	// Use reflection to call the specified method on the node
// 	node := t.nodes[evt.To]
// 	method := reflect.ValueOf(node).MethodByName(evt.Type)
// 	method.Call([]reflect.Value{
// 		reflect.ValueOf(evt.From),
// 		reflect.ValueOf(evt.To),
// 		reflect.ValueOf(evt.Value),
// 	})
// }

func (t *Simulator[T, S]) isCompleted() bool {
	// Is complete if all possible interleavings has been completed, i.e. all leaf nodes are "End" events
	if t.numRuns > 50000 {
		// If the simulation has run more than 50k runs we automatically stop it
		return true
	}
	return t.end
}
