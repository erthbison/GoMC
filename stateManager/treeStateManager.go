package stateManager

import (
	"fmt"
	"gomc/state"
	"gomc/tree"
	"io"
	"sync"

	"golang.org/x/exp/maps"
)

// Organizes the discovered StateSpace as a tree structure
//
// Collect the discovered runs as a tree with the initial state as the root.
// A path from the root to a leaf node is one run.
type TreeStateManager[T, S any] struct {
	sync.RWMutex
	stateRoot *tree.Tree[state.GlobalState[S]]

	getLocalState func(*T) S
	stateEq       func(S, S) bool
}

// Create a new TreeStateManager
//
// The TreeStateManager is configured with a function collecting the local state from a node
// and a function checking the equality of two states.
func NewTreeStateManager[T, S any](getLocalState func(*T) S, stateEq func(S, S) bool) *TreeStateManager[T, S] {
	return &TreeStateManager[T, S]{
		getLocalState: getLocalState,
		stateEq:       stateEq,
	}
}

// Adds the run to the discovered state space.
//
// Ïs safe to call from multiple goroutines.
func (sm *TreeStateManager[T, S]) AddRun(run []state.GlobalState[S]) {
	sm.Lock()
	defer sm.Unlock()

	if len(run) < 1 {
		return
	}

	currentTree := sm.stateRoot
	// If the tree has not been initialized:
	// Initialize it with the initial state as the root
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
