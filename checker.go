package gomc

import (
	"fmt"
	"gomc/tree"
)

type CheckerResponse interface {
	Response() (bool, string)
}

type predicateCheckerResponse[S, T any] struct {
	Result   bool                // True if all tests holds. False otherwise
	Sequence []GlobalState[S, T] // A sequence of states leading to the false test. nil if Result is true
	Test     int                 // The index of the failing test. -1 if Result is true
}

func (pcr predicateCheckerResponse[S, T]) Response() (bool, string) {
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

type PredicateChecker[S, T any] struct {
	// A slice of predicates that returns true if the predicate holds.
	// If the predicate is broken it returns false and a counterexample
	// The functions take the state and a boolean value indicating wether the node is a leaf node or not
	predicates []func(GlobalState[S, T], bool, []GlobalState[S, T]) bool
}

func NewPredicateChecker[S, T any](predicates ...func(GlobalState[S, T], bool, []GlobalState[S, T]) bool) *PredicateChecker[S, T] {
	return &PredicateChecker[S, T]{
		predicates: predicates,
	}
}

func (pc *PredicateChecker[S, T]) Check(root *tree.Tree[GlobalState[S, T]]) *predicateCheckerResponse[S, T] {
	// Checks that all predicates holds for all nodes. Nodes are searched depth first and the search is interrupted if some state that breaks the predicates are provided
	if resp := pc.checkNode(root, []GlobalState[S, T]{}); resp != nil {
		return resp
	}
	return &predicateCheckerResponse[S, T]{
		Result:   true,
		Sequence: nil,
		Test:     -1,
	}
}

func (pc *PredicateChecker[S, T]) checkNode(node *tree.Tree[GlobalState[S, T]], sequence []GlobalState[S, T]) *predicateCheckerResponse[S, T] {
	// Use a depth first search to search trough all nodes and check with predicates
	// Immediately stops when finding a state that breaches the predicates
	sequence = append(sequence, node.Payload())
	if ok, index := pc.checkState(node.Payload(), node.IsLeafNode(), sequence); !ok {
		return &predicateCheckerResponse[S, T]{
			Result:   false,
			Sequence: sequence,
			Test:     index,
		}
	}

	for _, child := range node.Children() {
		if resp := pc.checkNode(child, sequence); resp != nil {
			return resp
		}
	}
	return nil
}

func (pc *PredicateChecker[S, T]) checkState(state GlobalState[S, T], terminalState bool, sequence []GlobalState[S, T]) (bool, int) {
	// Check the state of a node on all predicates. The leafNode variable is true if this is a leaf node
	for index, pred := range pc.predicates {
		if !pred(state, terminalState, sequence) {
			return false, index
		}
	}
	return true, -1
}
