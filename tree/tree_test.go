package tree

import "testing"

func TestTreeAddChild(t *testing.T) {
	// Basic test to make sure that it works. Add some nodes and check some basic properties to ensure that they have been added correctly
	tree := New("Tree 1", func(a, b string) bool { return a == b })
	tree.AddChild("Tree 1-1")
	child := tree.AddChild("Tree 1-2")
	child.AddChild("Tree 1-2-1")

	if !tree.IsRoot() {
		t.Fatalf("Tree should be root node")
	}
	if tree.Len() != 4 {
		t.Fatalf("Added four elements to the tree. Has length: %v", tree.Len())
	}
	if len(tree.Children) != 2 {
		t.Fatalf("Added two children to the tree. Got: %v", len(tree.Children))
	}
	if child.IsRoot() {
		t.Fatalf("This should be a child node. IsRoot(): %v", child.IsRoot())
	}

	// Check that the Value has been added
	if !tree.DepthFirstSearch(func(s string) bool {
		return s == "Tree 1-2-1"
	}) {
		t.Fatalf("The value \"Tree 1-2-1\" should be a descendant of this node, but it cant be found with a depth first search")
	}

	if tree.SearchLeafNodes(func(s string) bool {
		return s == "Tree 1-2"
	}) {
		t.Fatalf("There is no element with value \"Tree 1-2\" in a leaf node")
	}

	if !tree.SearchLeafNodes(func(s string) bool {
		return s == "Tree 1-1"
	}) {
		t.Fatalf("There should be an element with value \"Tree 1-1\" in a leaf node")
	}
}
