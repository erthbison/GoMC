package tree

import (
	"fmt"
	"strings"
)

type Tree[T any] struct {
	payload  T
	parent   *Tree[T]
	children []*Tree[T]
	depth    int
	eq       func(a, b T) bool
}

func New[T any](payload T, eq func(a, b T) bool) Tree[T] {
	return Tree[T]{
		payload:  payload,
		parent:   nil,
		children: []*Tree[T]{},
		depth:    0,
		eq:       eq,
	}
}

// Returns the total number of elements in the tree
func (t *Tree[T]) Len() int {
	len := 1
	for _, child := range t.children {
		len += child.Len()
	}
	return len
}

// Adds a new child with the provided payload as a child of the current Tree
// Returns the child when done
func (t *Tree[T]) AddChild(payload T) *Tree[T] {
	treeNode := &Tree[T]{
		payload:  payload,
		parent:   t,
		children: []*Tree[T]{},
		depth:    t.depth + 1,
		eq:       t.eq,
	}
	t.children = append(t.children, treeNode)
	return treeNode
}

// Returns true if the TreeNode has a child with the provided payload.
// Otherwise returns false
func (t *Tree[T]) HasChild(payload T) bool {
	for _, node := range t.Children() {
		if t.eq(payload, node.Payload()) {
			return true
		}
	}
	return false
}

// Returns the first child node with the provided payload.
// If no such child node exists returns nil
func (t *Tree[T]) GetChild(payload T) *Tree[T] {
	for _, node := range t.Children() {
		if t.eq(payload, node.Payload()) {
			return node
		}
	}
	return nil
}

// String representation of a TreeNode
func (t *Tree[T]) String() string {
	out := strings.Builder{}
	for i := 0; i < t.Depth(); i++ {
		out.WriteString("-")
	}
	out.WriteString(fmt.Sprintf("%v\n", t.Payload()))
	for _, child := range t.Children() {
		out.WriteString(fmt.Sprintf("%v", child))
	}
	return out.String()
}

func (t *Tree[T]) IsRoot() bool {
	return t.Parent() == nil
}

func (t *Tree[T]) IsLeafNode() bool {
	return len(t.Children()) == 0
}

// Returns a slice of all tree nodes that are a descendent of this tree node
func (t *Tree[T]) GetAllLeafNodes() []*Tree[T] {
	leafNodes := []*Tree[T]{}
	if t.IsLeafNode() {
		leafNodes = append(leafNodes, t)
		return leafNodes
	}
	for _, child := range t.Children() {
		leafNodes = append(leafNodes, child.GetAllLeafNodes()...)
	}
	return leafNodes
}

// Returns true if the search function is true for some leaf node
func (t *Tree[T]) SearchLeafNodes(search func(T) bool) bool {
	if t.IsLeafNode() {
		if search(t.Payload()) {
			return true
		}
	}
	for _, child := range t.Children() {
		if child.SearchLeafNodes(search) {
			return true
		}
	}
	return false
}

// Returns true if the search function is true for some node
// Performs a DFS to find the node
func (t *Tree[T]) DepthFirstSearch(search func(T) bool) bool {
	if search(t.Payload()) {
		return true
	}
	for _, child := range t.Children() {
		if child.DepthFirstSearch(search) {
			return true
		}
	}
	return false
}

func (t *Tree[T]) Payload() T {
	return t.payload
}

func (t *Tree[T]) Parent() *Tree[T] {
	return t.parent
}

func (t *Tree[T]) Depth() int {
	return t.depth
}

func (t *Tree[T]) Children() []*Tree[T] {
	return t.children
}

func (t *Tree[T]) Newick() string {
	out := strings.Builder{}
	if len(t.Children()) > 0 {
		out.WriteString("(")
		for i, child := range t.Children() {
			if i > 0 {
				out.WriteString(",")
			}
			out.WriteString(child.Newick())
		}
		out.WriteString(")")
	}
	out.WriteString(fmt.Sprintf("\"%v\"", t.Payload()))
	if t.IsRoot() {
		out.WriteString(";")
	}
	return out.String()
}
