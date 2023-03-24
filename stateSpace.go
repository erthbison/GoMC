package gomc

import (
	"fmt"
	"gomc/tree"
	"io"
)

type StateSpace[S any] interface {
	Payload() GlobalState[S]
	Children() []StateSpace[S]
	IsTerminal() bool

	Export(io.Writer)
}

// A wrapper around the Tree structure so that it implements the StateSpace interface
type treeStateSpace[S any] struct {
	*tree.Tree[GlobalState[S]]
}

func (tss treeStateSpace[S]) Children() []StateSpace[S] {
	out := []StateSpace[S]{}
	for _, child := range tss.Tree.Children() {
		out = append(out, treeStateSpace[S]{
			Tree: child,
		})
	}
	return out
}

func (tss treeStateSpace[S]) IsTerminal() bool {
	return tss.IsLeafNode()
}

func (tss treeStateSpace[S]) Export(w io.Writer) {
	fmt.Fprint(w, tss.Newick())
}
