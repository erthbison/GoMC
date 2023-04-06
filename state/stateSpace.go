package state

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
type TreeStateSpace[S any] struct {
	*tree.Tree[GlobalState[S]]
}

func (tss TreeStateSpace[S]) Children() []StateSpace[S] {
	out := []StateSpace[S]{}
	for _, child := range tss.Tree.Children() {
		out = append(out, TreeStateSpace[S]{
			Tree: child,
		})
	}
	return out
}

func (tss TreeStateSpace[S]) IsTerminal() bool {
	return tss.IsLeafNode()
}

func (tss TreeStateSpace[S]) Export(w io.Writer) {
	fmt.Fprint(w, tss.Newick())
}
