package gomc

import (
	"gomc/tree"

	"golang.org/x/exp/maps"
)

type StateManager[T any, S any] interface {
	UpdateGlobalState(map[int]*T) // Update the state stored for this tick
	EndRun()                      // End the current run and prepare for the next
}

type stateManager[T any, S any] struct {
	StateRoot     *tree.Tree[map[int]S]
	currentState  *tree.Tree[map[int]S]
	getLocalState func(*T) S

	stateCmp func(S, S) bool
}

func NewStateManager[T any, S any](getLocalState func(*T) S, stateCmp func(S, S) bool) *stateManager[T, S] {
	stateRoot := tree.New(map[int]S{}, func(a, b map[int]S) bool {
		return maps.EqualFunc(a, b, stateCmp)
	})

	return &stateManager[T, S]{
		StateRoot:     &stateRoot,
		currentState:  &stateRoot,
		getLocalState: getLocalState,
		stateCmp:      stateCmp,
	}
}

func (sm *stateManager[T, S]) UpdateGlobalState(nodes map[int]*T) {
	states := map[int]S{}
	for id, node := range nodes {
		states[id] = sm.getLocalState(node)
	}

	// If the state already is a child of the current state, retrieve it and set it as the next state
	//  Otherwise add it as a child to the state tree
	if nextState := sm.currentState.GetChild(states); nextState != nil {
		sm.currentState = nextState
		return
	}
	sm.currentState = sm.currentState.AddChild(states)
}

func (sm *stateManager[T, S]) EndRun() {
	sm.currentState = sm.StateRoot
}
