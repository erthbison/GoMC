package state

import (
	"fmt"
	"gomc/tree"
	"io"
)

// A representation of the discovered State Space.
//
// The state space is represented as a set of nodes, each with references to their children.
// A path from the root of the state space to a terminal node should be a run of the algorithm.
type StateSpace[S any] interface {
	// Get the GlobalState stored in the node
	Payload() GlobalState[S]

	// Get the children of the current node.
	Children() []StateSpace[S]

	// Returns true if the state is the last state in a run.
	// Returns false otherwise.
	IsTerminal() bool

	// Export the state space to a writer.
	Export(io.Writer)
}

// A StateSpace that organizes the states in a tree structure.
//
// The root of the tree is the initial state of the system.
// The children of a node is the states that came after the current state in some of the runs.
// Exports the StateSpace as a Newick representation of the tree, with parenthesis around the payload.
// A wrapper around the Tree structure so that it implements the StateSpace interface
type TreeStateSpace[S any] struct {
	*tree.Tree[GlobalState[S]]
}

// Get the children of the current node.
func (tss TreeStateSpace[S]) Children() []StateSpace[S] {
	children := tss.Tree.Children()
	out := make([]StateSpace[S], len(children))
	for i, child := range children {
		out[i] = TreeStateSpace[S]{
			Tree: child,
		}
	}
	return out
}

// Returns true if the state is the last state in a run.
// Returns false otherwise.
func (tss TreeStateSpace[S]) IsTerminal() bool {
	return tss.IsLeafNode()
}

// Export the state space to a writer.
//
// Exports the StateSpace as a Newick representation of the tree, with parenthesis around the payload.
func (tss TreeStateSpace[S]) Export(w io.Writer) {
	fmt.Fprint(w, tss.Newick())
}
