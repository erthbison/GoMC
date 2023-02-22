package predicate

import "gomc"

// Check that the predicate happens eventually.
// This is done by only running the predicate on the terminalState of all sequences
// For all non-terminal states the predicate always returns true
func Eventually[S any](pred func(gomc.GlobalState[S], []gomc.GlobalState[S]) bool) gomc.Predicate[S] {
	return func(states gomc.GlobalState[S], terminalState bool, sequence []gomc.GlobalState[S]) bool {
		if !terminalState {
			return true
		}
		return pred(states, sequence)
	}
}

// Check that the condition is true for all processes in the provided global state
// If checkCorrect is true, only correct processes will be checked, otherwise faulty processes will also be checked.
func ForAllNodes[S any](cond func(S) bool, gs gomc.GlobalState[S], checkCorrect bool) bool {
	for id, state := range gs.LocalStates {
		if checkCorrect && !gs.Correct[id] {
			continue
		}
		if !cond(state) {
			return false
		}
	}
	return true
}
