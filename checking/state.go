package checking

import "gomc/state"

type State[S any] struct {
	LocalStates map[int]S
	Correct     map[int]bool
	IsTerminal  bool
	Sequence    []state.GlobalState[S]
}
