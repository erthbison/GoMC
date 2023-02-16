package gomc

import (
	"fmt"
	"gomc/event"
	"gomc/tree"

	"golang.org/x/exp/maps"
)

type StateManager[T any, S any] interface {
	UpdateGlobalState(map[int]*T, map[int]bool, event.Event) // Update the state stored for this tick
	EndRun()                                                 // End the current run and prepare for the next
}

type GlobalState[S any] struct {
	LocalStates map[int]S
	Correct     map[int]bool
	Evt         event.Event
}

func (gs GlobalState[S]) String() string {
	crashed := []int{}
	for id, status := range gs.Correct {
		if !status {
			crashed = append(crashed, id)
		}
	}
	return fmt.Sprintf("Evt: %v|States: %v|Crashed: %v", gs.Evt, gs.LocalStates, crashed)
}

type stateManager[T any, S any] struct {
	StateRoot     *tree.Tree[GlobalState[S]]
	currentState  *tree.Tree[GlobalState[S]]
	getLocalState func(*T) S

	stateCmp func(S, S) bool
}

func NewStateManager[T any, S any](getLocalState func(*T) S, stateCmp func(S, S) bool) *stateManager[T, S] {
	stateRoot := tree.New(GlobalState[S]{}, func(a, b GlobalState[S]) bool {
		if !event.EventsEquals(a.Evt, b.Evt) {
			return false
		}
		if !maps.EqualFunc(a.LocalStates, b.LocalStates, stateCmp) {
			return false
		}
		return maps.Equal(a.Correct, b.Correct)
	})

	return &stateManager[T, S]{
		StateRoot:     &stateRoot,
		currentState:  &stateRoot,
		getLocalState: getLocalState,
		stateCmp:      stateCmp,
	}
}

func (sm *stateManager[T, S]) UpdateGlobalState(nodes map[int]*T, correct map[int]bool, evt event.Event) {
	states := map[int]S{}
	for id, node := range nodes {
		states[id] = sm.getLocalState(node)
	}

	copiedCorrect := map[int]bool{}
	for id, status := range correct {
		copiedCorrect[id] = status
	}
	globalState := GlobalState[S]{
		LocalStates: states,
		Correct:     copiedCorrect,
		Evt:         evt,
	}

	// If the state already is a child of the current state, retrieve it and set it as the next state
	//  Otherwise add it as a child to the state tree
	if nextState := sm.currentState.GetChild(globalState); nextState != nil {
		sm.currentState = nextState
		return
	}
	sm.currentState = sm.currentState.AddChild(globalState)
}

func (sm *stateManager[T, S]) EndRun() {
	sm.currentState = sm.StateRoot
}
