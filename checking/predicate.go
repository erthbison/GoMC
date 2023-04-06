package checking

// Check that the predicate happens eventually.
// This is done by only running the predicate on the terminalState of all sequences
// For all non-terminal states the predicate always returns true
func Eventually[S any](pred Predicate[S]) Predicate[S] {
	return func(s State[S]) bool {
		if !s.IsTerminal {
			return true
		}
		return pred(s)
	}
}

// Check that the condition is true for all processes in the provided global state
// If checkCorrect is true, only correct processes will be checked, otherwise faulty processes will also be checked.
func ForAllNodes[S any](cond func(S) bool, s State[S], checkCorrect bool) bool {
	for id, state := range s.LocalStates {
		if checkCorrect && !s.Correct[id] {
			continue
		}
		if !cond(state) {
			return false
		}
	}
	return true
}
