package tester

import (
	"errors"
	"fmt"
	"log"
	"reflect"
)

var (
	NoMessagesError = errors.New("tester: No messages in the queue matching the event")
	UnknownEvent    = errors.New("tester: Received unknown event type")
)

/*
	Requirements of nodes:
		- Must call the Send function with the required input when sending a message.
		- The message type in the Send function must correspond to a method of the node that takes the input (int, int, []byte)
*/

type Simulator[T any, S any] struct {
	nodes map[int]*T

	// Responsibility for maintaining the state space.
	sm StateManager[T, S]

	// The event tree should be a tree of events that has been discovered during the traversal of the state space
	// It includes both paths that have been fully explored, and potential paths that we know need further interleaving of the messages to fully explore
	Scheduler Scheduler

	end     bool
	numRuns int
}

func CreateSimulator[T any, S any](sch Scheduler, sm StateManager[T, S]) *Simulator[T, S] {
	return &Simulator[T, S]{
		nodes:     map[int]*T{},
		Scheduler: sch,
		sm:        sm,
		numRuns:   0,
		end:       false,
	}
}

func (t *Simulator[T, S]) Simulate(initNodes func() map[int]*T, start func(map[int]*T)) {
	// t.getLocalState = getLocalState
	for !t.isCompleted() {
		// Create nodes and init states for this run
		t.nodes = initNodes()
		t.sm.UpdateGlobalState(t.nodes)

		start(t.nodes)
		for {
			err := t.executeNextEvent()
			if err != nil {
				if errors.Is(err, NoEventError) {
					// If there are no available events that means that all possible event chains have been attempted and we are done
					// Update the end flag
					t.end = true
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
	t.Scheduler.AddEvent(Event{
		Type: "Message",
		Payload: Message{
			From:  from,
			To:    to,
			Type:  msgType,
			Value: msg,
		},
	})
}

func (t *Simulator[T, S]) executeNextEvent() error {
	event, err := t.Scheduler.GetEvent()
	if err != nil {
		return err
	}
	switch event.Type {
	case "Message":
		t.sendMessage(event.Payload)
	default:
		return UnknownEvent
	}
	return nil
}

func (t *Simulator[T, S]) sendMessage(msg Message) {
	// Use reflection to call the specified method on the node
	node := t.nodes[msg.To]
	method := reflect.ValueOf(node).MethodByName(msg.Type)
	method.Call([]reflect.Value{
		reflect.ValueOf(msg.From),
		reflect.ValueOf(msg.To),
		reflect.ValueOf(msg.Value),
	})
}

func (t *Simulator[T, S]) isCompleted() bool {
	// Is complete if all possible interleavings has been completed, i.e. all leaf nodes are "End" events
	if t.numRuns > 50000 {
		// If the simulation has run more than 50k runs we automatically stop it
		return true
	}
	return t.end
}
