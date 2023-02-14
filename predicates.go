package gomc

func PredEventually[S, T any](pred func(GlobalState[S, T], bool, []GlobalState[S, T]) bool) func(GlobalState[S, T], bool, []GlobalState[S, T]) bool {
	return func(states GlobalState[S, T], terminalState bool, sequence []GlobalState[S, T]) bool {
		if !terminalState {
			return true
		}
		return pred(states, terminalState, sequence)
	}
}
