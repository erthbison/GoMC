package checking

// Check that the predicate happens eventually.
//
// Return a predicate that run the provided predicate on terminal states.
// Returns the value of the original predicate if the state is terminal.
// Otherwise, it always returns true.
func Eventually[S any](pred Predicate[S]) Predicate[S] {
	return func(s State[S]) bool {
		if !s.IsTerminal {
			return true
		}
		return pred(s)
	}
}

// Check that condition returns true for all processes in the provided global state
//
// Returns false if cond returns false for some node.
// Returns true otherwise.
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
