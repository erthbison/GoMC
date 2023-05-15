package stateManager

import (
	"gomc/event"
	"gomc/state"

	"golang.org/x/exp/maps"
)

// A type that manages the state of a single run at a time.
//
// Should only be accessed from a single goroutine at a time.
// When the run has been completed and the EndRun function is called the run is sent on the send channel and the state is reset.
// The RunStateManager can then safely be used on a new run.
type RunStateManager[T, S any] struct {
	sm            StateManager[T, S]
	getLocalState func(*T) S

	run []state.GlobalState[S]
}

// Create a new RunStateManager
//
// Is initialized with a reference to the StateManager that created it
// and a function specifying how to collect the state from the node.
func NewRunStateManager[T, S any](sm StateManager[T, S], getLocalState func(*T) S) *RunStateManager[T, S] {
	return &RunStateManager[T, S]{
		sm:            sm,
		getLocalState: getLocalState,

		run: make([]state.GlobalState[S], 0),
	}
}

// Collect the state from the nodes and add the GlobalState to the current run
// 
// Aggregate the local states, the status, and the event into the GlobalState and add it to the run. 
// nodes is the map of nodes used in this run.
// correct is a map of the status of the nodes.
// evt is the event that caused the transition into the current state.
func (rss *RunStateManager[T, S]) UpdateGlobalState(nodes map[int]*T, correct map[int]bool, evt event.Event) {
	states := map[int]S{}
	for id, node := range nodes {
		states[id] = rss.getLocalState(node)
	}

	rss.run = append(rss.run, state.GlobalState[S]{
		LocalStates: states,
		Correct:     maps.Clone(correct),
		Evt:         state.CreateEventRecord(evt),
	})
}

func (rss *RunStateManager[T, S]) EndRun() {
	rss.sm.AddRun(rss.run)
	rss.run = make([]state.GlobalState[S], 0)
}
