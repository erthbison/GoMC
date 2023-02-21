package predicate

import "gomc"

func Eventually[S any](pred gomc.Predicate[S]) gomc.Predicate[S] {
	return func(states gomc.GlobalState[S], terminalState bool, sequence []gomc.GlobalState[S]) bool {
		if !terminalState {
			return true
		}
		return pred(states, terminalState, sequence)
	}
}
