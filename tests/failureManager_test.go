package gomc_test

import (
	"gomc"
	"testing"
)

func TestFailureManager(t *testing.T) {
	fm := gomc.NewFailureManager()
	nodes := []int{0, 1, 2, 3, 4}
	fm.Init(nodes)
	for _, node := range nodes {
		if !fm.CorrectNodes()[node] {
			t.Errorf("Expected all nodes to be correct. %v is not", node)
		}
	}
	called := false
	callbackFunc := func(nodeId int) {
		called = true
		if nodeId != 4 {
			t.Errorf("Expected node 4 to fail")
		}
	}
	fm.Subscribe(callbackFunc)

	fm.NodeCrash(4)
	if !called {
		t.Errorf("Expected the provided callback function to be called")
	}

	fm.Init(nodes)
	for _, node := range nodes {
		if !fm.CorrectNodes()[node] {
			t.Errorf("Expected all nodes to be correct. %v is not", node)
		}
	}
	fm.NodeCrash(4)
	if !called {
		t.Errorf("Expected the provided callback function to be called")
	}
}
