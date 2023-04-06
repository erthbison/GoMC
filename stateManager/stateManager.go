package stateManager

import (
	"gomc/state"
	"gomc/tree"
	"sync"
)

// Manages the global state across several runs.
type StateManager[T, S any] interface {
	GetRunStateManager() *RunStateManager[T, S]
	AddRun(run []state.GlobalState[S])
	State() state.StateSpace[S]
}

// A type that manages the state of several runs of the same system and represent all discovered states of the system.
// The TreeStateManger waits for completed runs on the send channel. Once they are received they are added to the state tree that is used to represent the state space.
// After calling Stop() the send channel will be closed and the TreeStateManager will no longer add new runs to the state tree
type TreeStateManager[T, S any] struct {
	sync.RWMutex
	stateRoot *tree.Tree[state.GlobalState[S]]

	getLocalState func(*T) S
	stateEq       func(S, S) bool
}
