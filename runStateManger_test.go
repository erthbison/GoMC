package gomc

import (
	"fmt"
	"sync"
	"testing"
)

func TestStateMangerMerge(t *testing.T) {
	for i, test := range mergeTest {
		sm := NewTreeStateManager(
			func(t *MockNode) int { return t.Id },
			func(i1, i2 int) bool { return i1 == i2 },
		)
		inChan := make(chan []int)
		var wait sync.WaitGroup
		wait.Add(len(test.runs))
		for j := 0; j < test.numProcesses; j++ {
			go func(numNodes, i int) {
				for runLength := range inChan {
					rst := sm.NewRun()
					for _, k := range runLength{
						nodes, correct, evt := generateMockData(numNodes, k)
						rst.UpdateGlobalState(nodes, correct, evt)
					}
					rst.EndRun()
					wait.Done()
				}
			}(test.numNodes, i)
		}
		for _, runLength := range test.runs {
			inChan <- runLength
		}
		wait.Wait()
		close(inChan)
		sm.Stop()

		if sm.StateRoot.Len() != test.expectedLen {
			fmt.Println(sm.StateRoot.Newick())
			t.Errorf("Test %v: Unexpected Size of the state tree. Got %v. Expected %v", i, sm.StateRoot.Len(), test.expectedLen)
		}
	}
}

func TestTreeStateManagerEndRun(t *testing.T) {
	for i, test := range stateMangerTest {
		outChan := make(chan []GlobalState[int], 1)
		sm := &RunStateManager[MockNode, int]{
			run:           make([]GlobalState[int], 0),
			getLocalState: func(t *MockNode) int { return t.Id },
			send:          outChan,
		}
		for _, runLength := range test.runs {
			for j := 0; j < runLength; j++ {
				nodes, correct, evt := generateMockData(test.numNodes, j)
				sm.UpdateGlobalState(nodes, correct, evt)
			}
			sm.EndRun()
			receivedRun := <-outChan
			if len(receivedRun) != runLength {
				t.Errorf("Test %v Failed: Received unexpected run. Got run of length %v. Wanted %v", i, len(receivedRun), runLength)
			}
		}
	}
}

// Generate mock data for the run. The content of the data is not important, what is important is that it is properly stored
func generateMockData(numNodes, i int) (map[int]*MockNode, map[int]bool, MockEvent) {
	nodes := map[int]*MockNode{}
	correct := map[int]bool{}
	for id := 0; id < numNodes; id++ {
		nodes[id] = &MockNode{Id: i}
		correct[id] = true
	}
	evt := MockEvent{i + numNodes, 0, false}
	return nodes, correct, evt
}

var stateMangerTest = []struct {
	numNodes int
	runs     []int // A slice of the length of each run
}{
	{
		numNodes: 0, runs: []int{},
	},
	{
		numNodes: 1, runs: []int{},
	},
	{
		numNodes: 0, runs: []int{1},
	},
	{
		numNodes: 1, runs: []int{10},
	},
	{
		numNodes: 1, runs: []int{10, 10},
	},
}

var mergeTest = []struct {
	numProcesses int
	numNodes     int
	runs         [][]int // A slice of the length of each run
	expectedLen  int
}{
	{
		numProcesses: 1,
		numNodes:     3,
		runs:         [][]int{{0, 1, 2, 3, 4}, {0, 1, 2, 3, 4}},
		expectedLen:  5,
	},
	{
		numProcesses: 3,
		numNodes:     3,
		runs:         [][]int{{}, {}, {0}},
		expectedLen:  1,
	},
	{
		numProcesses: 5,
		numNodes:     3,
		runs:         [][]int{{}, {}, {0}},
		expectedLen:  1,
	},
	{
		numProcesses: 5,
		numNodes:     3,
		runs:         [][]int{{0, 1, 2, 3, 4}, {0, 1, 2, 7, 8}, {0, 1, 2, 9, 10}},
		expectedLen:  9,
	},
}
