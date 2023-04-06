package gomc

import (
	"gomc/event"
)

// A type that manages the state of a single run at a time.
// Should only be accessed from a single goroutine at a time.
// When the run has been completed and the EndRun function is called the run is sent on the send channel and the state is reset.
// The RunStateManager can then safely be used on a new run.
type RunStateManager[T, S any] struct {
	sm            StateManager[T, S]
	getLocalState func(*T) S

	run []GlobalState[S]
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
	rss.run = append(rss.run, GlobalState[S]{
		LocalStates: states,
		Correct:     copiedCorrect,
		Evt:         createEventRecord(evt),
	})
}

func (rss *RunStateManager[T, S]) EndRun() {
	rss.sm.AddRun(rss.run)
	rss.run = make([]GlobalState[S], 0)
}
