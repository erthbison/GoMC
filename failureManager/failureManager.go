package failureManager

import "errors"

// The failureManager keeps track of which nodes has crashed and which has not.
// It also provides a Subscribe(func(int)) function which can be used to emulate the properties of a perfect failure detector
// The subscribe function replicates the functionality of a perfect failure detectors.
// All provided callback functions are called immediately upon the crash of a node
type FailureManager struct {
	correct         map[int]bool
	failureCallback []func(int)
}

func New() *FailureManager {
	return &FailureManager{
		correct:         make(map[int]bool),
		failureCallback: make([]func(int), 0),
	}
}

// Init the failure manager with the provided nodes.
// Set all the provided nodes to correct
func (fm *FailureManager) Init(nodes []int) {
	for _, id := range nodes {
		fm.correct[id] = true
	}
}

// Return a map of the status of the nodes
func (fm *FailureManager) CorrectNodes() map[int]bool {
	return fm.correct
}

// register that the node with the provided id has crashed
func (fm *FailureManager) NodeCrash(nodeId int) error {
	if _, ok := fm.correct[nodeId]; !ok {
		return errors.New("FailureManager: Received NodeCrash for node that is not added to the system")
	}
	fm.correct[nodeId] = false
	for _, f := range fm.failureCallback {
		f(nodeId)
	}
	return nil
}

// Register a callback function to be called when a node crashes.
func (fm *FailureManager) Subscribe(callback func(int)) {
	fm.failureCallback = append(fm.failureCallback, callback)
}
