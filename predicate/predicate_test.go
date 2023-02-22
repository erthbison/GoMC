package predicate

import (
	"gomc"
	"testing"
)

func TestEventually(t *testing.T) {
	emptySeq := make([]gomc.GlobalState[bool], 0)
	eventuallyPred := func(gs1 gomc.GlobalState[bool], b bool, gs2 []gomc.GlobalState[bool]) bool {
		for _, n := range gs1.LocalStates {
			if !n {
				return false
			}
		}
		return true
	}
	for i, test := range eventuallyTest {
		pred := Eventually(eventuallyPred)
		out := pred(test.gs, test.terminal, emptySeq)
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
		out := ForAllNodes(cond, test.gs, test.checkCorrect)
		if out != test.expected {
			t.Errorf("Received unexpected bool from predicate on test %v. Got %v", i, out)
		}
	}

}

var eventuallyTest = []struct {
	terminal bool
	gs       gomc.GlobalState[bool]
	expected bool
}{
	{
		false,
		gomc.GlobalState[bool]{},
		true,
	},
	{
		true,
		gomc.GlobalState[bool]{LocalStates: map[int]bool{0: true, 1: true}},
		true,
	},
	{
		true,
		gomc.GlobalState[bool]{LocalStates: map[int]bool{0: true, 1: false}},
		false,
	},
	{
		false,
		gomc.GlobalState[bool]{LocalStates: map[int]bool{0: true, 1: false}},
		true,
	},
}

var forAllNodesTest = []struct {
	gs           gomc.GlobalState[bool]
	checkCorrect bool
	expected     bool
}{
	{
		gs: gomc.GlobalState[bool]{
			LocalStates: map[int]bool{0: true, 1: true, 2: true},
			Correct:     map[int]bool{0: true, 1: true, 2: true},
		},
		checkCorrect: true,
		expected:     true,
	},
	{
		gs: gomc.GlobalState[bool]{
			LocalStates: map[int]bool{0: false, 1: true, 2: true},
			Correct:     map[int]bool{0: true, 1: true, 2: true},
		},
		checkCorrect: false,
		expected:     false,
	},
	{
		gs: gomc.GlobalState[bool]{
			LocalStates: map[int]bool{0: false, 1: true, 2: true},
			Correct:     map[int]bool{0: false, 1: true, 2: true},
		},
		checkCorrect: false,
		expected:     false,
	},
	{
		gs: gomc.GlobalState[bool]{
			LocalStates: map[int]bool{0: false, 1: true, 2: true},
			Correct:     map[int]bool{0: false, 1: true, 2: true},
		},
		checkCorrect: true,
		expected:     true,
	},
}
