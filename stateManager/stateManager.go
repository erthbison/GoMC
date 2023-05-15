package stateManager

import (
	"gomc/state"
)

// Manages the global state across several runs.
type StateManager[T, S any] interface {
	// Create a RunStateManager to be used to simulate a run.
	//
	// The RunStateManager will collect the state of a run and report it back to the StateManager
	GetRunStateManager() *RunStateManager[T, S]

	// Add a run to the runs collected during this simulation.
	//
	// Will be called from multiple goroutines
	AddRun(run []state.GlobalState[S])

	// Collect the StateSpace that was discovered during the simulation.
	State() state.StateSpace[S]

	// Reset the StateSpace and prepare for a new simulation.
	Reset()
}
