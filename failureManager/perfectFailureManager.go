package failureManager

import (
	"errors"
	"gomc/event"
	"gomc/eventManager"
)

// The PerfectFailureManager is a failure manager that implements the PerfectFailureDetector abstraction in a fail-stop system.
//
// It is configured with a slice of nodes that will crash at some point during the simulation, i.e. nodes that are faulty.
// it is also configured with a function specifying how the node should crash.
type PerfectFailureManager[T any] struct {
	crashFunc    func(*T)
	failingNodes []int
}

// Create a new PerfectFailureManager
//
// Implements the PerfectFailureDetector abstraction in a fail-stop system.
// crashFunc is a function performing the crash on the node.
// It should close all network connections and stop all ongoing executions on the node.
// Events executed on the node after the crash should have no effect.
// failingNodes is a slice of node ids of the nodes that will crash at some point during a run.
func NewPerfectFailureManager[T any](crashFunc func(*T), failingNodes []int) *PerfectFailureManager[T] {
	return &PerfectFailureManager[T]{
		crashFunc:    crashFunc,
		failingNodes: failingNodes,
	}
}

// Create a RunFailureManager that can be used when simulating a run
// ea is the EventAdder that is used in this run.
// The EventAdder for the run is provided in the SimulationParameters
func (pfm PerfectFailureManager[T]) GetRunFailureManager(ea eventManager.EventAdder) RunFailureManager[T] {
	return newRunPerfectFailureManager(ea, pfm.crashFunc, pfm.failingNodes)
}

// The run specific implementation of the PerfectFailureManager
//
// Manages the functionality of the PerfectFailureManager during the simulation of a run.
type runPerfectFailureManager[T any] struct {
	ea           eventManager.EventAdder
	crashFunc    func(*T)
	failingNodes []int

	correct         map[int]bool
	nodes           map[int]*T
	failureCallback map[int]func(int, bool)
}

// Create a new runPerfectFailureManager
//
// ea is the EventAdder that is used in this run
// crashFunc is a function performing the crash on the node.
// failingNodes is a slice of node ids of the nodes that will crash at some point during a run.
func newRunPerfectFailureManager[T any](ea eventManager.EventAdder, crashFunc func(*T), failingNodes []int) *runPerfectFailureManager[T] {
	return &runPerfectFailureManager[T]{
		ea:           ea,
		crashFunc:    crashFunc,
		failingNodes: failingNodes,

		correct:         make(map[int]bool),
		failureCallback: make(map[int]func(int, bool)),
	}
}

// Initialize the FailureManager with the nodes that are used in this run
func (fm *runPerfectFailureManager[T]) Init(nodes map[int]*T) {
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
			event.NewCrashEvent(id, fm.nodeCrash),
		)
	}
}

// Return a map of the node ids and the status of the corresponding node
//
// If the status is true the node is currently running.
// if it is false the node has crashed.
func (fm *runPerfectFailureManager[T]) CorrectNodes() map[int]bool {
	return fm.correct
}

// Perform the crash of the node with the provided id.
//
// The method is called by the CrashEvent when it is executed.
func (fm *runPerfectFailureManager[T]) nodeCrash(nodeId int) error {
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

// Subscribe to updates about node status.
//
// id is the id of the node that subscribes to the callback.
// The callback is a function that is called with the new status of the node when the status.
func (fm *runPerfectFailureManager[T]) Subscribe(id int, callback func(int, bool)) {
	fm.failureCallback[id] = callback
}
