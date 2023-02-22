package gomc

import (
	"fmt"
	"gomc/event"
	"gomc/tree"
	"io"

	"golang.org/x/exp/maps"
)

type StateManager[T any, S any] interface {
	UpdateGlobalState(map[int]*T, map[int]bool, event.Event) // Update the state stored for this tick
	EndRun()                                                 // End the current run and prepare for the next
	Export(io.Writer)                                        // A function to export the state space to a writer
}

type GlobalState[S any] struct {
	LocalStates map[int]S    // A map storing the local state of the nodes. The map stores (id, state) combination.
	Correct     map[int]bool // A map storing the status of the node. The map stores (id, status) combination. If status is true, the node with id "id" is correct, otherwise the node is true. All nodes are represented in the map.
	evt         event.Event  // A copy of the event that caused the transition into this state. It should not be changed.
}

func (gs GlobalState[S]) String() string {
	crashed := []int{}
	for id, status := range gs.Correct {
		if !status {
			crashed = append(crashed, id)
		}
	}
	return fmt.Sprintf("Evt: %v\t States: %v\t Crashed: %v\t", gs.evt, gs.LocalStates, crashed)
}

type treeStateManager[T any, S any] struct {
	StateRoot     *tree.Tree[GlobalState[S]]
	currentState  *tree.Tree[GlobalState[S]]
	getLocalState func(*T) S

	stateCmp func(S, S) bool
}

func NewTreeStateManager[T any, S any](getLocalState func(*T) S, stateCmp func(S, S) bool) *treeStateManager[T, S] {
	stateRoot := tree.New(GlobalState[S]{}, func(a, b GlobalState[S]) bool {
		if !event.EventsEquals(a.evt, b.evt) {
			return false
		}
		if !maps.EqualFunc(a.LocalStates, b.LocalStates, stateCmp) {
			return false
		}
		return maps.Equal(a.Correct, b.Correct)
	})

	return &treeStateManager[T, S]{
		StateRoot:     &stateRoot,
		currentState:  &stateRoot,
		getLocalState: getLocalState,
		stateCmp:      stateCmp,
	}
}

// Retrieve the new global state and store it
func (sm *treeStateManager[T, S]) UpdateGlobalState(nodes map[int]*T, correct map[int]bool, evt event.Event) {
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
		evt:         evt,
	}

	// If the state already is a child of the current state, retrieve it and set it as the next state
	//  Otherwise add it as a child to the state tree
	if nextState := sm.currentState.GetChild(globalState); nextState != nil {
		sm.currentState = nextState
		return
	}
	sm.currentState = sm.currentState.AddChild(globalState)
}

// Mark a run as ended and returns to the state root to begin the next run
func (sm *treeStateManager[T, S]) EndRun() {
	sm.currentState = sm.StateRoot
}

// Write the Newick representation of the state tree to the writer
func (sm *treeStateManager[T, S]) Export(wrt io.Writer) {
	fmt.Fprint(wrt, sm.StateRoot.Newick())
}
