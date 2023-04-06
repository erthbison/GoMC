package checking

import "gomc/state"

// The state of the system at the current point of execution
type State[S any] struct {
	LocalStates map[int]S              // The local states of the nodes.
	Correct     map[int]bool           // The status of the nodes. True means that the node is correct, false that it has crashed.
	IsTerminal  bool                   // True if this is the last recorded state in a run. False otherwise.
	Sequence    []state.GlobalState[S] // The sequence of GlobalStates that lead to this State.
}
