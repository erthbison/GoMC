package tester

import (
	"experimentation/tree"
	"fmt"
)

type Checker[S any] interface {
	Check(*tree.Tree[map[int]S]) CheckerResponse
}
type CheckerResponse interface {
	Response() (bool, string)
}

type predicateCheckerResponse[S any] struct {
	Result   bool        // True if all tests holds. False otherwise
	Sequence []map[int]S // A sequence of states leading to the false test. nil if Result is true
	Test     int         // The index of the failing test. -1 if Result is true
}

func (pcr predicateCheckerResponse[S]) Response() (bool, string) {
	if pcr.Result {
		return pcr.Result, "All predicates holds"
	}
	out := fmt.Sprintf("Predicate broken. Predicate: %v. Sequence: \n", pcr.Test)
	for _, element := range pcr.Sequence {
		out += fmt.Sprintln("->", element)
	}
	return pcr.Result, out
}

// TODO: Consider if we can define the predicates as tests and automatically discover them

// TODO: Consider generalizing this so that it does not depend on the tree structure, but instead can work on some arbitrary data structure

type PredicateChecker[S any] struct {
	// A slice of predicates that returns true if the predicate holds.
	// If the predicate is broken it returns false and a counterexample
	// The functions take the state and a boolean value indicating wether the node is a leaf node or not
	predicates []func(map[int]S, bool) bool
}

func NewPredicateChecker[S any](preds ...func(map[int]S, bool) bool) *PredicateChecker[S] {
	return &PredicateChecker[S]{
		predicates: preds,
	}
}

func (pc *PredicateChecker[S]) Check(root *tree.Tree[map[int]S]) *predicateCheckerResponse[S] {
	// Checks that all predicates holds for all nodes. Nodes are searched depth first and the search is interrupted if some state that breaks the predicates are provided
	if resp := pc.checkNode(root); resp != nil {
		return resp
	}
	return &predicateCheckerResponse[S]{
		Result:   true,
		Sequence: nil,
		Test:     -1,
	}
}

func (pc *PredicateChecker[S]) checkNode(node *tree.Tree[map[int]S]) *predicateCheckerResponse[S] {
	// Use a depth first search to search trough all nodes and check with predicates
	// Immediately stops when finding a state that breaches the predicates
	if ok, index := pc.checkState(node.Payload, node.IsLeafNode()); !ok {
		return &predicateCheckerResponse[S]{
			Result:   false,
			Sequence: pc.getSequence(node),
			Test:     index,
		}
	}

	for _, child := range node.Children {
		if resp := pc.checkNode(child); resp != nil {
			return resp
		}
	}
	return nil
}

func (pc *PredicateChecker[S]) checkState(state map[int]S, terminalState bool) (bool, int) {
	// Check the state of a node on all predicates. The leafNode variable is true if this is a leaf node
	for index, pred := range pc.predicates {
		if !pred(state, terminalState) {
			return false, index
		}
	}
	return true, -1
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
