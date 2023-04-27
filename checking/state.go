package checking

import "gomc/state"

// The state of the system at the current point of execution
type State[S any] struct {
	// The local states of the nodes.
	LocalStates map[int]S
	// The status of the nodes. True means that the node is correct, false that it has crashed.
	Correct map[int]bool
	// True if this is the last recorded state in a run. False otherwise.
	IsTerminal bool
	// The sequence of GlobalStates that lead to this State.
	Sequence []state.GlobalState[S]
}
