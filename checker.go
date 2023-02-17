package gomc

import (
	"bytes"
	"fmt"
	"gomc/tree"
	"text/tabwriter"
)

type CheckerResponse interface {
	Response() (bool, string)
	Export() []string
}

type predicateCheckerResponse[S any] struct {
	Result      bool             // True if all tests holds. False otherwise
	Sequence    []GlobalState[S] // A sequence of states leading to the false test. nil if Result is true
	Test        int              // The index of the failing test. -1 if Result is true
	EvtSequence []string
}

func (pcr predicateCheckerResponse[S]) Response() (bool, string) {
	if pcr.Result {
		return pcr.Result, "All predicates holds"
	}
	var buffer bytes.Buffer
	wrt := tabwriter.NewWriter(&buffer, 4, 4, 0, ' ', 0)
	out := fmt.Sprintf("Predicate broken. Predicate: %v. Sequence: \n", pcr.Test)
	for _, element := range pcr.Sequence {
		fmt.Fprintf(wrt, "-> %v \n", element)
	}
	wrt.Flush()
	out += buffer.String()
	return pcr.Result, out
}

func (pcr predicateCheckerResponse[S]) Export() []string {
	return pcr.EvtSequence
}

// TODO: Consider if we can define the predicates as tests and automatically discover them

// TODO: Consider generalizing this so that it does not depend on the tree structure, but instead can work on some arbitrary data structure

type PredicateChecker[S any] struct {
	// A slice of predicates that returns true if the predicate holds.
	// If the predicate is broken it returns false and a counterexample
	// The functions take the state and a boolean value indicating wether the node is a leaf node or not
	predicates []func(GlobalState[S], bool, []GlobalState[S]) bool
}

func NewPredicateChecker[S any](predicates ...func(GlobalState[S], bool, []GlobalState[S]) bool) *PredicateChecker[S] {
	return &PredicateChecker[S]{
		predicates: predicates,
	}
}

func (pc *PredicateChecker[S]) Check(root *tree.Tree[GlobalState[S]]) *predicateCheckerResponse[S] {
	// Checks that all predicates holds for all nodes. Nodes are searched depth first and the search is interrupted if some state that breaks the predicates are provided
	if resp := pc.checkNode(root, []GlobalState[S]{}); resp != nil {
		return resp
	}
	return &predicateCheckerResponse[S]{
		Result:      true,
		Sequence:    nil,
		Test:        -1,
		EvtSequence: nil,
	}
}

func (pc *PredicateChecker[S]) checkNode(node *tree.Tree[GlobalState[S]], sequence []GlobalState[S]) *predicateCheckerResponse[S] {
	// Use a depth first search to search trough all nodes and check with predicates
	// Immediately stops when finding a state that breaches the predicates
	sequence = append(sequence, node.Payload())
	if ok, index := pc.checkState(node.Payload(), node.IsLeafNode(), sequence); !ok {
		evtSequence := []string{}
		for _, state := range sequence {
			if state.Evt != nil {
				evtSequence = append(evtSequence, state.Evt.Id())
			}
		}
		return &predicateCheckerResponse[S]{
			Result:      false,
			Sequence:    sequence,
			Test:        index,
			EvtSequence: evtSequence,
		}
	}

	for _, child := range node.Children() {
		if resp := pc.checkNode(child, sequence); resp != nil {
			return resp
		}
	}
	return nil
}

func (pc *PredicateChecker[S]) checkState(state GlobalState[S], terminalState bool, sequence []GlobalState[S]) (bool, int) {
	// Check the state of a node on all predicates. The leafNode variable is true if this is a leaf node
	for index, pred := range pc.predicates {
		if !pred(state, terminalState, sequence) {
			return false, index
		}
	}
	return true, -1
}
