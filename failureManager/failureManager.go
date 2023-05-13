package failureManager

import "gomc/eventManager"

// Used to manage the correctness of nodes
//
// Keeps track of which nodes has crashed.
// Imitates the functionality of a failure manager and provide mechanisms that allows nodes to learn which nodes have crashed.
// Performs changes in the status of nodes.
//
// Configures the RunFailureManager and tracks global information across runs
type FailureManger[T any] interface {
	// Create a RunFailureManager that can be used when simulating a run
	GetRunFailureManager(eventManager.EventAdder) RunFailureManager[T]
}

// RunSpecific part of the FailureManager
//
// Manages the functionality of the Failure Manager across a single run.
type RunFailureManager[T any] interface {
	// Initialize the FailureManager with the nodes that are used in this run
	Init(nodes map[int]*T)

	// Return a map of the node ids and the status of the corresponding node
	//
	// If the status is true the node is currently running.
	// if it is false the node has crashed.
	CorrectNodes() map[int]bool

	// Subscribe to updates about node status.
	//
	// id is the id of the node that subscribes to the callback.
	// The callback is a function that is called with the new status of the node when the status.
	// A node should only subscribe to node crashes once.
	Subscribe(id int, callback func(id int, status bool))
}
