package checking

import (
	"bytes"
	"fmt"
	"gomc/event"
	"gomc/state"
	"text/tabwriter"
)

type predicateCheckerResponse[S any] struct {
	Result   bool                   // True if all tests holds. False otherwise
	Sequence []state.GlobalState[S] // A sequence of states leading to the false test. nil if Result is true
	Test     int                    // The index of the failing test. -1 if Result is true
}

// Generate a response
// Returns two parameters, result, and description.
// Result is true if all predicates hold, false otherwise.
// Description is a formatted string providing a detailed description of the result.
// If result is false the description contain a representation of the sequence of states that lead to the failing state
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

// Export the failing event sequence to a slice of strings to be replayed by the ReplayScheduler
func (pcr predicateCheckerResponse[S]) Export() []event.EventId {
	evtSequence := []event.EventId{}
	if pcr.Sequence == nil {
		return evtSequence
	}
	for _, state := range pcr.Sequence {
		evtSequence = append(evtSequence, state.Evt.Id)
	}
	return evtSequence
}

// A function to be evaluated on the states
// It returns true if the predicate holds for the state and false otherwise
type Predicate[S any] func(s State[S]) bool

type PredicateChecker[S any] struct {
	// A slice of predicates that returns true if the predicate holds.
	// If the predicate is broken it returns false and a counterexample
	// The functions take the state and a boolean value indicating wether the node is a leaf node or not
	predicates []Predicate[S]
}

func NewPredicateChecker[S any](predicates ...Predicate[S]) *PredicateChecker[S] {
	return &PredicateChecker[S]{
		predicates: predicates,
	}
}

func (pc *PredicateChecker[S]) Check(root state.StateSpace[S]) *predicateCheckerResponse[S] {
	// Checks that all predicates holds for all nodes. Nodes are searched depth first and the search is interrupted if some state that breaks the predicates are provided
	if resp := pc.checkNode(root, []state.GlobalState[S]{}); resp != nil {
		return resp
	}
	return &predicateCheckerResponse[S]{
		Result:   true,
		Sequence: nil,
		Test:     -1,
	}
}

func (pc *PredicateChecker[S]) checkNode(node state.StateSpace[S], sequence []state.GlobalState[S]) *predicateCheckerResponse[S] {
	// Use a depth first search to search trough all nodes and check with predicates
	// Immediately stops when finding a state that breaches the predicates
	sequence = append(sequence, node.Payload())
	if ok, index := pc.checkState(node.Payload(), node.IsTerminal(), sequence); !ok {
		return &predicateCheckerResponse[S]{
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

func (pc *PredicateChecker[S]) checkState(state state.GlobalState[S], terminalState bool, sequence []state.GlobalState[S]) (bool, int) {
	// Check the state of a node on all predicates. The leafNode variable is true if this is a leaf node
	for index, pred := range pc.predicates {
		if !pred(State[S]{
			LocalStates: state.LocalStates,
			Correct:     state.Correct,
			IsTerminal:  terminalState,
			Sequence:    sequence,
		}) {
			return false, index
		}
	}
	return true, -1
}
