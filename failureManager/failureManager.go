package failureManager

// Used to manage the correctness of nodes
type FailureManager interface {
	Init(nodeIds []int)                           // Initialize the FailureManager with a set of nodes for this run
	CorrectNodes() map[int]bool                   // Return a map of the node ids and the status of the corresponding node
	NodeCrash(nodeId int) error                   // Perform the crash of a node
	Subscribe(callback func(id int, status bool)) // Subscribe to updates about node status. Calls the callback function with the node id and the new status of the node when the status of a node changes
}
