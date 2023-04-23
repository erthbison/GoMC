package failureManager

import (
	"gomc/event"
	"testing"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func TestInit(t *testing.T) {
	for i, test := range InitTest {
		sch := NewMockRunScheduler()
		fm := newPerfectRunFailureManager(
			sch, func(t *MockNode) { t.crashed = true }, test.failingNodes,
		)
		fm.Init(test.nodes)

		correct := fm.CorrectNodes()
		expectedCorrect := map[int]bool{}
		for id := range test.nodes {
			expectedCorrect[id] = true
		}

		if !maps.Equal(expectedCorrect, correct) {
			t.Errorf("Test %v: Correct nodes has not been initialized correctly. Expected: %v. Got: %v", i, expectedCorrect, correct)
		}

		addedCrashes := []int{}
		for _, evt := range sch.addedEvents {
			addedCrashes = append(addedCrashes, evt.Target())
		}

		if !slices.Equal(test.scheduledCrashes, addedCrashes) {
			t.Errorf("Test %v: Incorrect events scheduled. Expected: %v, Got: %v", i, test.scheduledCrashes, addedCrashes)
		}

	}
}

func TestNodeCrash(t *testing.T) {
	for i, test := range NodeCrashTest {
		sch := NewMockRunScheduler()
		fm := newPerfectRunFailureManager(
			sch, func(t *MockNode) { t.crashed = true }, []int{},
		)
		fm.nodes = test.nodes
		fm.correct = test.correct

		// For testing crash function
		expectedNodeCrashed := map[int]bool{}
		for id, node := range test.nodes {
			expectedNodeCrashed[id] = node.crashed
		}
		if _, ok := test.nodes[test.crashingNode]; ok {
			expectedNodeCrashed[test.crashingNode] = true
		}

		fm.Subscribe(test.subscribingNodeId, func(nodeId int, status bool) {})

		err := fm.NodeCrash(test.crashingNode)

		// Test that an error is returned as expected
		isErr := (err != nil)
		if isErr != test.expectedErr {
			if isErr {
				t.Errorf("Test %v: Expected no error got: %v", i, err)
			} else {
				t.Errorf("Test %v: Expected to receive an error", i)
			}
		}

		// Test that CorrectNodes() is properly updated
		// Clone the correct map before calling NodeCrash, then make the expected change
		expectedCorrect := maps.Clone(test.correct)
		if _, ok := test.correct[test.crashingNode]; ok {
			expectedCorrect[test.crashingNode] = false
		}

		if !maps.Equal(expectedCorrect, fm.CorrectNodes()) {
			t.Errorf("Test %v: Unexpected map of correct nodes. Expected: %v, Got: %v", i, expectedCorrect, fm.CorrectNodes())
		}

		// Test that the provided crash function is properly called
		actualNodeCrashed := map[int]bool{}
		for id, node := range test.nodes {
			actualNodeCrashed[id] = node.crashed
		}

		if !maps.Equal(expectedNodeCrashed, actualNodeCrashed) {
			t.Errorf("Test %v: Crash function has been called on unexpected node. Expected: %v, Got: %v", i, expectedNodeCrashed, actualNodeCrashed)
		}

		// Test that the failure callbacks are properly called
		if !test.expectedErr {
			evt := sch.addedEvents[0]
			cd, ok := evt.(event.CrashDetection)
			if !ok {
				t.Errorf("Test %v: Expected CrashDetection event to have been added to scheduler", i)
			}
			if cd.Target() != test.subscribingNodeId {
				t.Errorf("Test %v: Expected CrashDetection event to target node 0", i)
			}
		} else {
			if len(sch.addedEvents) != 0 {
				t.Errorf("Test %v: Expected no event to have been added to the scheduler.", i)
			}
		}

	}
}

var InitTest = []struct {
	// The provided slice of node id for which we will try to schedule crashes
	// May contain invalid values
	failingNodes []int
	// The nodes which we will init
	nodes map[int]*MockNode

	// The node id for which a crash is expected to be scheduled
	scheduledCrashes []int
}{
	{
		[]int{},
		map[int]*MockNode{0: {}, 1: {}, 2: {}},

		[]int{},
	},
	{
		[]int{0, 1},
		map[int]*MockNode{0: {}, 1: {}, 2: {}},

		[]int{0, 1},
	},

	{
		[]int{0, 10},
		map[int]*MockNode{0: {}, 1: {}, 2: {}},

		[]int{0},
	},

	{
		[]int{10},
		map[int]*MockNode{0: {}, 1: {}, 2: {}},

		[]int{},
	},
}

var NodeCrashTest = []struct {
	nodes             map[int]*MockNode
	correct           map[int]bool
	subscribingNodeId int
	crashingNode      int
	expectedErr       bool
}{
	{
		map[int]*MockNode{0: {}, 1: {}, 2: {}},
		map[int]bool{0: true, 1: true, 2: true},
		3,
		0,
		false,
	},
	{
		map[int]*MockNode{0: {}, 1: {}, 2: {}},
		map[int]bool{0: true, 1: true, 2: true},
		3,
		5,
		true,
	},
	{
		map[int]*MockNode{0: {}, 1: {crashed: true}, 2: {}},
		map[int]bool{0: true, 1: false, 2: true},
		3,
		1,
		true,
	},
}
