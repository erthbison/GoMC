package tester

func PredEventually[S any](pred func(map[int]S, bool, []map[int]S) bool) func(map[int]S, bool, []map[int]S) bool {
	return func(states map[int]S, terminalState bool, sequence []map[int]S) bool {
		if !terminalState {
			return true
		}
		return pred(states, terminalState, sequence)
	}
}
