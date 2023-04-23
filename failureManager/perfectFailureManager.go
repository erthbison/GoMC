package failureManager

import (
	"errors"
	"gomc/event"
	"gomc/eventManager"
)

// The failureManager keeps track of which nodes has crashed and which has not.
// It also provides a Subscribe(func(int)) function which can be used to emulate the properties of a perfect failure detector
// The subscribe function replicates the functionality of a perfect failure detectors.
// All provided callback functions are called immediately upon the crash of a node
type PerfectFailureManager[T any] struct {
	crashFunc    func(*T)
	failingNodes []int
}

func NewPerfectFailureManager[T any](crashFunc func(*T), failingNodes []int) *PerfectFailureManager[T] {
	return &PerfectFailureManager[T]{
		crashFunc:    crashFunc,
		failingNodes: failingNodes,
	}
}

func (pfm PerfectFailureManager[T]) GetRunFailureManager(ea eventManager.EventAdder) RunFailureManager[T] {
	return newPerfectRunFailureManager(ea, pfm.crashFunc, pfm.failingNodes)
}

type PerfectRunFailureManager[T any] struct {
	ea           eventManager.EventAdder
	crashFunc    func(*T)
	failingNodes []int

	correct         map[int]bool
	nodes           map[int]*T
	failureCallback map[int]func(int, bool)
}

func newPerfectRunFailureManager[T any](ea eventManager.EventAdder, crashFunc func(*T), failingNodes []int) *PerfectRunFailureManager[T] {
	return &PerfectRunFailureManager[T]{
		ea:           ea,
		crashFunc:    crashFunc,
		failingNodes: failingNodes,

		correct:         make(map[int]bool),
		failureCallback: make(map[int]func(int, bool)),
	}
}

// Init the failure manager with the provided nodes.
// Set all the provided nodes to correct
func (fm *PerfectRunFailureManager[T]) Init(nodes map[int]*T) {
	for id := range nodes {
		fm.correct[id] = true
	}

	fm.nodes = nodes

	// Schedule crash events
	for _, id := range fm.failingNodes {
		if _, ok := nodes[id]; !ok {
			continue
		}
		fm.ea.AddEvent(
			event.NewCrashEvent(id, fm.NodeCrash),
		)
	}
}

// Return a map of the status of the nodes
func (fm *PerfectRunFailureManager[T]) CorrectNodes() map[int]bool {
	return fm.correct
}

// register that the node with the provided id has crashed
func (fm *PerfectRunFailureManager[T]) NodeCrash(nodeId int) error {
	node, ok := fm.nodes[nodeId]
	if !ok {
		return errors.New("FailureManager: Received NodeCrash for node that is not added to the system")
	}

	if status := fm.correct[nodeId]; !status {
		return errors.New("FailureManager: Received NodeCrash for node that has already crashed. Is failStop abstraction so node can not crash again.")
	}
	// Set node as crashed
	fm.correct[nodeId] = false

	// Call the provided crash function with the node
	fm.crashFunc(node)

	// Call all provided crash callbacks
	for id, f := range fm.failureCallback {
		fm.ea.AddEvent(event.NewCrashDetection(
			id,
			nodeId,
			f,
		))
	}
	return nil
}

// Register a callback function to be called when a node crashes.
func (fm *PerfectRunFailureManager[T]) Subscribe(id int, callback func(int, bool)) {
	fm.failureCallback[id] = callback
}
