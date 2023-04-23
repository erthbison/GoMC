package runner

import (
	"testing"
	"time"
)

var (
	eventBuffer = 1000
	getState    = func(n *MockNode) int { return n.val }
	crashFunc   = func(n *MockNode) { n.crashed = true }
)

func TestNodeControllerMain(t *testing.T) {
	for i, test := range mainTest {
		recordChan := make(chan Record)
		nc := NewNodeController(0, test.node, getState, crashFunc, recordChan, eventBuffer)

		for _, evt := range test.events {
			nc.addEvent(evt)
		}

		go nc.Main()

		currentState := 0
		for j := 0; j < len(test.events)*2; j++ {
			rec := <-recordChan
			if state, ok := rec.(StateRecord[int]); ok {
				expectedState := test.expectedStateChange[currentState]
				if state.state != expectedState {
					t.Errorf("Test %v: Events executed in unexpected order. Got state: %v. Expected: %v", i, state.state, expectedState)
				}
				currentState++
			}
		}

		select {
		case rec := <-recordChan:
			t.Errorf("Test %v: No more record expected got: %v", i, rec)
		default:
		}

	}
}

func TestNodeControllerPause(t *testing.T) {
	recordChan := make(chan Record)
	node := &MockNode{}
	nc := NewNodeController(0, node, getState, crashFunc, recordChan, eventBuffer)

	go nc.Main()

	nc.Pause()

	nc.addEvent(MockEvent{val: 10})

	// Nc is paused
	// Check that no record is received and that the state does not change
	select {
	case rec := <-recordChan:
		t.Errorf("NodeController paused. Did not expect to get value: %v", rec)
	default:
	}

	if node.val != 0 {
		t.Errorf("NodeController paused. Did not expect the state of the node to be updated")
	}

	// Resume execution
	nc.Resume()

	// check that event is recorded
	select {
	case rec := <-recordChan:
		if _, ok := rec.(ExecutionRecord); !ok {
			t.Errorf("Expected to get an ExecutionRecord. Got %T", rec)
		}
	case <-time.After(5 * time.Second):
		t.Errorf("NodeController resumed. Did expect to get value. Got none")
	}

	// Check that state is recorded
	select {
	case rec := <-recordChan:
		state, ok := rec.(StateRecord[int])
		if !ok {
			t.Errorf("Expected to get an StateRecord. Got %T", rec)
		}
		if state.state != 10 {
			t.Errorf("NodeController resumed. Expected to receive state record with correct value. Got: %v. Expected: %v", state.state, 10)
		}
	case <-time.After(5 * time.Second):
		t.Errorf("NodeController resumed. Did expect to get value. Got none")
	}

	if node.val != 10 {
		t.Errorf("NodeController resumed. Did expect the state of the node to be updated. Got: %v", node.val)
	}
}

func TestClose(t *testing.T) {
	recordChan := make(chan Record)
	node := &MockNode{}
	nc := NewNodeController(0, node, getState, crashFunc, recordChan, eventBuffer)

	stopped := make(chan bool)
	go func() {
		nc.Main()
		stopped <- true
	}()

	nc.Close()

	nc.addEvent(MockEvent{val: 10})

	select {
	case rec := <-recordChan:
		t.Errorf("NodeController stopped. Did not expect to receive an event. Got %T", rec)
	case <-time.After(100 * time.Millisecond):
	}

	select {
	case <-stopped:
	default:
		t.Errorf("Expected main loop to have stopped")
	}
}

func TestCloseAfterPause(t *testing.T) {
	recordChan := make(chan Record)
	node := &MockNode{}
	nc := NewNodeController(0, node, getState, crashFunc, recordChan, eventBuffer)

	stopped := make(chan bool)
	go func() {
		nc.Main()
		stopped <- true
	}()
	nc.Pause()

	nc.Close()

	nc.addEvent(MockEvent{val: 10})

	select {
	case rec := <-recordChan:
		t.Errorf("NodeController stopped. Did not expect to receive an event. Got %T", rec)
	case <-time.After(100 * time.Millisecond):
	}

	select {
	case <-stopped:
	default:
		t.Errorf("Expected main loop to have stopped")
	}
}

func TestAllowDuplicateClose(t *testing.T) {
	recordChan := make(chan Record)
	node := &MockNode{}
	nc := NewNodeController(0, node, getState, crashFunc, recordChan, eventBuffer)

	stopped := make(chan bool)
	go func() {
		nc.Main()
		stopped <- true
	}()

	nc.Close()

	nc.Pause()

	nc.Close()

	nc.Resume()

	nc.Close()

	nc.addEvent(MockEvent{val: 10})

	select {
	case rec := <-recordChan:
		t.Errorf("NodeController stopped. Did not expect to receive an event. Got %T", rec)
	case <-time.After(100 * time.Millisecond):
	}

	select {
	case <-stopped:
	default:
		t.Errorf("Expected main loop to have stopped")
	}
}

func TestResumeBeforePause(t *testing.T) {
	recordChan := make(chan Record)
	node := &MockNode{}
	nc := NewNodeController(0, node, getState, crashFunc, recordChan, eventBuffer)

	go func() {
		nc.Main()
	}()

	nc.Resume()

	nc.Pause()

	nc.Pause()

	nc.Pause()

	nc.Resume()

	stopped := make(chan bool)
	go func() {
		nc.Main()
		stopped <- true
	}()

	nc.addEvent(MockEvent{val: 10})

	// check that event is recorded
	select {
	case rec := <-recordChan:
		if _, ok := rec.(ExecutionRecord); !ok {
			t.Errorf("Expected to get an ExecutionRecord. Got %T", rec)
		}
	case <-time.After(5 * time.Second):
		t.Errorf("NodeController resumed. Did expect to get value. Got none")
	}

	// Check that state is recorded
	select {
	case rec := <-recordChan:
		state, ok := rec.(StateRecord[int])
		if !ok {
			t.Errorf("Expected to get an StateRecord. Got %T", rec)
		}
		if state.state != 10 {
			t.Errorf("NodeController resumed. Expected to receive state record with correct value. Got: %v. Expected: %v", state.state, 10)
		}
	case <-time.After(5 * time.Second):
		t.Errorf("NodeController resumed. Did expect to get value. Got none")
	}

	if node.val != 10 {
		t.Errorf("NodeController resumed. Did expect the state of the node to be updated. Got: %v", node.val)
	}
}

var mainTest = []struct {
	node   *MockNode
	events []MockEvent

	expectedStateChange []int
}{
	{
		&MockNode{},
		[]MockEvent{{val: 5}, {val: 10}},

		[]int{5, 10},
	},
	{
		&MockNode{},
		[]MockEvent{},

		[]int{},
	},
	{
		&MockNode{},
		[]MockEvent{{val: 5}, {val: 10}, {val: 10}},

		[]int{5, 10, 10},
	},
}
