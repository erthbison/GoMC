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

	err := fm.NodeCrash(4)
	if err != nil {
		t.Errorf("Did not expect to receive an error. Got %v", err)
	}
	if !called {
		t.Errorf("Expected the provided callback function to be called")
	}

	fm.Init(nodes)
	for _, node := range nodes {
		if !fm.CorrectNodes()[node] {
			t.Errorf("Expected all nodes to be correct. %v is not", node)
		}
	}
	err = fm.NodeCrash(4)
	if err != nil {
		t.Errorf("Did not expect to receive an error. Got %v", err)
	}
	if !called {
		t.Errorf("Expected the provided callback function to be called")
	}
}

func TestRandomId(t *testing.T) {
	fm := gomc.NewFailureManager()
	nodes := []int{1, 485, 786, 354, 458, 456}
	fm.Init(nodes)
	for _, node := range nodes {
		if !fm.CorrectNodes()[node] {
			t.Errorf("Expected all nodes to be correct. %v is not", node)
		}
	}
	called := false
	callbackFunc := func(nodeId int) {
		called = true
		if nodeId != 354 {
			t.Errorf("Expected node 354 to fail. %v failed instead.", nodeId)
		}
	}
	fm.Subscribe(callbackFunc)

	err := fm.NodeCrash(354)
	if err != nil {
		t.Errorf("Did not expect to receive an error. Got %v", err)
	}
	if !called {
		t.Errorf("Expected the provided callback function to be called")
	}

	fm.Init(nodes)
	for _, node := range nodes {
		if !fm.CorrectNodes()[node] {
			t.Errorf("Expected all nodes to be correct. %v is not", node)
		}
	}
	err = fm.NodeCrash(354)
	if err != nil {
		t.Errorf("Did not expect to receive an error. Got %v", err)
	}
	if !called {
		t.Errorf("Expected the provided callback function to be called")
	}

	err = fm.NodeCrash(0)
	if err == nil {
		t.Errorf("Provided invalid nodeId. Expected to receive an error")
	}
}
