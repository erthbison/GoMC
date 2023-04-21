package stateManager

import (
	"gomc/event"
	"gomc/state"
	"os"
	"strconv"
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
					rst := sm.GetRunStateManager()
					for _, k := range runLength {
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
		state := sm.State().(state.TreeStateSpace[int])
		size := state.Len()
		if size != test.expectedLen {
			state.Export(os.Stdout)
			t.Errorf("Test %v: Unexpected Size of the state tree. Got %v. Expected %v", i, size, test.expectedLen)
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
	evt := MockEvent{event.EventId(strconv.Itoa(i)), 0, false, 0}
	return nodes, correct, evt
}

var mergeTest = []struct {
	numProcesses int
	numNodes     int
	runs         [][]int // A slice of runs, where a run is represented by a slice of the ordered ids in the run
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
		runs:         [][]int{{}, {0}, {0}},
		expectedLen:  1,
	},
	{
		numProcesses: 5,
		numNodes:     3,
		runs:         [][]int{{}, {0}, {0}},
		expectedLen:  1,
	},
	{
		numProcesses: 5,
		numNodes:     3,
		runs:         [][]int{{0, 1, 2, 3, 4}, {0, 1, 2, 7, 8}, {0, 1, 2, 9, 10}},
		expectedLen:  9,
	},
}
