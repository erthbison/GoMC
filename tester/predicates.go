package tester

func PredEventually[S any](pred func(map[int]S, bool) bool) func(map[int]S, bool) bool {
	return func(states map[int]S, terminalState bool) bool {
		if !terminalState {
			return true
		}
		return pred(states, terminalState)
	}
}
