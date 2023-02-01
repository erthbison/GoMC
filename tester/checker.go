package tester

import (
	"experimentation/tree"
)

type Checker[S any] interface {
	Check(*tree.Tree[map[int]S]) CheckerResponse[S]
}

type CheckerResponse[S any] struct {
	Result   bool        // True if all tests holds. False otherwise
	Sequence []map[int]S // A sequence of states leading to the false test. nil if Result is true
	Test     string      // A description of the failing test. Empty string if Result is true
}

// TODO: Consider if we can define the predicates as tests and automatically discover them

type PredicateChecker[S any] struct {
	// A slice of predicates that returns true if the predicate holds.
	// If the predicate is broken it returns false and a counterexample
	// The functions take the state and a boolean value indicating wether the node is a leaf node or not
	predicates []func(map[int]S, bool) (bool, string)
}

func NewPredicateChecker[S any](preds ...func(map[int]S, bool) (bool, string)) *PredicateChecker[S] {
	return &PredicateChecker[S]{
		predicates: preds,
	}
}

func (pc *PredicateChecker[S]) Check(root *tree.Tree[map[int]S]) *CheckerResponse[S] {
	// Checks that all predicates holds for all nodes. Nodes are searched depth first and the search is interrupted if some state that breaks the predicates are provided
	if resp := pc.checkNode(root); resp != nil {
		return resp
	}
	return &CheckerResponse[S]{
		Result:   true,
		Sequence: nil,
		Test:     "",
	}
}

func (pc *PredicateChecker[S]) checkNode(node *tree.Tree[map[int]S]) *CheckerResponse[S] {
	// Use a depth first search to search trough all nodes and check with predicates
	// Immediately stops when finding a state that breaches the predicates
	if ok, desc := pc.checkState(node.Payload, node.IsLeafNode()); !ok {
		return &CheckerResponse[S]{
			Result:   false,
			Sequence: pc.getSequence(node),
			Test:     desc,
		}
	}

	for _, child := range node.Children {
		if resp := pc.checkNode(child); resp != nil {
			return resp
		}
	}
	return nil
}

func (pc *PredicateChecker[S]) checkState(state map[int]S, leafNode bool) (bool, string) {
	// Check the state of a node on all predicates. The leafNode variable is true if this is a leaf node
	for _, pred := range pc.predicates {
		if ok, desc := pred(state, leafNode); !ok {
			return false, desc
		}
	}
	return true, ""
}

func (pc *PredicateChecker[S]) getSequence(node *tree.Tree[map[int]S]) []map[int]S {
	// Get the sequence leading to the provided nodes by traversing the tree upwards until it reaches the root
	sequence := make([]map[int]S, node.Depth+1)
	for i := node.Depth; i >= 0; i-- {
		sequence[i] = node.Payload
		node = node.Parent
	}
	return sequence
}
