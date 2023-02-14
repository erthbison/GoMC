package gomc

func PredEventually[S any](pred func(GlobalState[S], bool, []GlobalState[S]) bool) func(GlobalState[S], bool, []GlobalState[S]) bool {
	return func(states GlobalState[S], terminalState bool, sequence []GlobalState[S]) bool {
		if !terminalState {
			return true
		}
		return pred(states, terminalState, sequence)
	}
}
