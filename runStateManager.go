package gomc

import (
	"gomc/event"
	"gomc/state"
)

// A type that manages the state of a single run at a time.
// Should only be accessed from a single goroutine at a time.
// When the run has been completed and the EndRun function is called the run is sent on the send channel and the state is reset.
// The RunStateManager can then safely be used on a new run.
type RunStateManager[T, S any] struct {
	sm            StateManager[T, S]
	getLocalState func(*T) S

	run []state.GlobalState[S]
}

func (rss *RunStateManager[T, S]) UpdateGlobalState(nodes map[int]*T, correct map[int]bool, evt event.Event) {
	states := map[int]S{}
	for id, node := range nodes {
		states[id] = rss.getLocalState(node)
	}

	copiedCorrect := map[int]bool{}
	for id, status := range correct {
		copiedCorrect[id] = status
	}
	rss.run = append(rss.run, state.GlobalState[S]{
		LocalStates: states,
		Correct:     copiedCorrect,
		Evt:         state.CreateEventRecord(evt),
	})
}

func (rss *RunStateManager[T, S]) EndRun() {
	rss.sm.AddRun(rss.run)
	rss.run = make([]state.GlobalState[S], 0)
}
