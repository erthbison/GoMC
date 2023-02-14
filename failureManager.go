package gomc

// The failureManager keeps track of which nodes has crashed and which has not.
// It also provides a Subscribe(func(int)) function which can be used to emulate the properties of a perfect failure detector
type failureManager struct {
	correct         map[int]bool
	failureCallback []func(int)
}

func NewFailureManager() *failureManager {
	return &failureManager{
		correct:         make(map[int]bool),
		failureCallback: make([]func(int), 0),
	}
}

// Init the failure manager with the provided nodes.
// Set all the provided nodes to correct
func (fm *failureManager) Init(nodes []int) {
	for id := range nodes {
		fm.correct[id] = true
	}
}

// Return a map of the status of the nodes
func (fm *failureManager) CorrectNodes() map[int]bool {
	return fm.correct
}

// register that the node with the provided id has crashed
func (fm *failureManager) NodeCrash(nodeId int) {
	fm.correct[nodeId] = false
	for _, f := range fm.failureCallback {
		f(nodeId)
	}
}

// Register a callback function to be called when a node crashes.
func (fm *failureManager) Subscribe(callback func(int)) {
	fm.failureCallback = append(fm.failureCallback, callback)
}
