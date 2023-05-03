package stateManager

import (
	"fmt"
	"gomc/state"
	"gomc/tree"
	"io"
	"sync"

	"golang.org/x/exp/maps"
)

// A type that manages the state of several runs of the same system and represent all discovered states of the system.
// The TreeStateManger waits for completed runs on the send channel. Once they are received they are added to the state tree that is used to represent the state space.
// After calling Stop() the send channel will be closed and the TreeStateManager will no longer add new runs to the state tree
type TreeStateManager[T, S any] struct {
	sync.RWMutex
	stateRoot *tree.Tree[state.GlobalState[S]]

	getLocalState func(*T) S
	stateEq       func(S, S) bool
}

func NewTreeStateManager[T, S any](getLocalState func(*T) S, stateEq func(S, S) bool) *TreeStateManager[T, S] {
	return &TreeStateManager[T, S]{
		getLocalState: getLocalState,
		stateEq:       stateEq,
	}
}

func (sm *TreeStateManager[T, S]) AddRun(run []state.GlobalState[S]) {
	sm.Lock()
	defer sm.Unlock()
	currentTree := sm.stateRoot
	if len(run) < 1 {
		return
	}
	if currentTree == nil {
		currentTree = sm.initStateTree(run[0])
		sm.stateRoot = currentTree
	}
	for _, state := range run[1:] {
		// If the state already is a child of the current state, retrieve it and set it as the next state
		if nextState := currentTree.GetChild(state); nextState != nil {
			currentTree = nextState
			continue
		}
		// Otherwise add it as a child to the state tree
		currentTree = currentTree.AddChild(state)
	}
}

// Initializes the state tree with the provided state as the initial state
func (sm *TreeStateManager[T, S]) initStateTree(s state.GlobalState[S]) *tree.Tree[state.GlobalState[S]] {
	cmp := func(a, b state.GlobalState[S]) bool {
		if a.Evt.Id != b.Evt.Id {
			return false
		}
		if !maps.EqualFunc(a.LocalStates, b.LocalStates, sm.stateEq) {
			return false
		}
		return maps.Equal(a.Correct, b.Correct)
	}
	stateRoot := tree.New(s, cmp)
	return stateRoot
}

// Create a RunStateManager to be used to collect the state of the new run
func (sm *TreeStateManager[T, S]) GetRunStateManager() *RunStateManager[T, S] {
	return NewRunStateManager[T, S](sm, sm.getLocalState)
}

// Write the Newick representation of the state tree to the writer
func (sm *TreeStateManager[T, S]) Export(wrt io.Writer) {
	sm.RLock()
	defer sm.RUnlock()
	fmt.Fprint(wrt, sm.stateRoot.Newick())
}

func (sm *TreeStateManager[T, S]) State() state.StateSpace[S] {
	sm.RLock()
	defer sm.RUnlock()
	return state.TreeStateSpace[S]{Tree: sm.stateRoot}
}

func (sm *TreeStateManager[T, S]) Reset() {
	sm.RLock()
	defer sm.RUnlock()
	sm.stateRoot = nil
}
