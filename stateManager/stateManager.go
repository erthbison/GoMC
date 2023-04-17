package stateManager

import (
	"gomc/state"
)

// Manages the global state across several runs.
type StateManager[T, S any] interface {
	GetRunStateManager() *RunStateManager[T, S]
	AddRun(run []state.GlobalState[S])
	State() state.StateSpace[S]
}
