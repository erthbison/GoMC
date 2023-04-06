package checking

import (
	"gomc/state"
	"testing"
)

var emptySeq = make([]state.GlobalState[bool], 0)

func TestEventually(t *testing.T) {

	eventuallyPred := func(s State[bool]) bool {
		for _, n := range s.LocalStates {
			if !n {
				return false
			}
		}
		return true
	}
	for i, test := range eventuallyTest {
		pred := Eventually(eventuallyPred)
		s := State[bool]{
			LocalStates: test.gs.LocalStates,
			Correct:     test.gs.Correct,
			IsTerminal:  test.terminal,
			Sequence:    emptySeq,
		}
		out := pred(s)
		if out != test.expected {
			t.Errorf("Received unexpected bool from predicate on test %v. Got %v", i, out)
		}
	}
}

func TestForAllNodes(t *testing.T) {
	cond := func(s bool) bool {
		return s
	}
	for i, test := range forAllNodesTest {
		s := State[bool]{
			LocalStates: test.gs.LocalStates,
			Correct:     test.gs.Correct,
			IsTerminal:  false,
			Sequence:    emptySeq,
		}
		out := ForAllNodes(cond, s, test.checkCorrect)
		if out != test.expected {
			t.Errorf("Received unexpected bool from predicate on test %v. Got %v", i, out)
		}
	}

}

var eventuallyTest = []struct {
	terminal bool
	gs       state.GlobalState[bool]
	expected bool
}{
	{
		false,
		state.GlobalState[bool]{},
		true,
	},
	{
		true,
		state.GlobalState[bool]{LocalStates: map[int]bool{0: true, 1: true}},
		true,
	},
	{
		true,
		state.GlobalState[bool]{LocalStates: map[int]bool{0: true, 1: false}},
		false,
	},
	{
		false,
		state.GlobalState[bool]{LocalStates: map[int]bool{0: true, 1: false}},
		true,
	},
}

var forAllNodesTest = []struct {
	gs           state.GlobalState[bool]
	checkCorrect bool
	expected     bool
}{
	{
		gs: state.GlobalState[bool]{
			LocalStates: map[int]bool{0: true, 1: true, 2: true},
			Correct:     map[int]bool{0: true, 1: true, 2: true},
		},
		checkCorrect: true,
		expected:     true,
	},
	{
		gs: state.GlobalState[bool]{
			LocalStates: map[int]bool{0: false, 1: true, 2: true},
			Correct:     map[int]bool{0: true, 1: true, 2: true},
		},
		checkCorrect: false,
		expected:     false,
	},
	{
		gs: state.GlobalState[bool]{
			LocalStates: map[int]bool{0: false, 1: true, 2: true},
			Correct:     map[int]bool{0: false, 1: true, 2: true},
		},
		checkCorrect: false,
		expected:     false,
	},
	{
		gs: state.GlobalState[bool]{
			LocalStates: map[int]bool{0: false, 1: true, 2: true},
			Correct:     map[int]bool{0: false, 1: true, 2: true},
		},
		checkCorrect: true,
		expected:     true,
	},
}
