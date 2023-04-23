package failureManager

import "gomc/eventManager"

// Used to manage the correctness of nodes
type FailureManger[T any] interface {
	GetRunFailureManager(eventManager.EventAdder) RunFailureManager[T]
}

type RunFailureManager[T any] interface {
	Init(nodes map[int]*T)                                // Initialize the FailureManager with a set of nodes for this run
	CorrectNodes() map[int]bool                           // Return a map of the node ids and the status of the corresponding node
	Subscribe(id int, callback func(id int, status bool)) // Subscribe to updates about node status. Calls the callback function with the node id and the new status of the node when the status of a node changes
}

/*
	Failure Manager Should:
		- Keep track of which nodes has crashed and which has not
		- Provide mechanism that allows nodes to learn which nodes have crashed and which have not.
			- Specific should vary depending on abstraction
			- Should at least be able to support perfect FD
			- Should be able to support delayed information and node specific(i.e. not all nodes are informed at the same time)
				- This should probably be handled by using events.
		- Perform crashing, since the failure manager can interact with the nodes


*/
