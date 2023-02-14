package gomc

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

func (fm *failureManager) Init(nodes []int) {
	for id := range nodes {
		fm.correct[id] = true
	}
}

func (fm *failureManager) CorrectNodes() map[int]bool {
	return fm.correct
}

func (fm *failureManager) NodeCrash(nodeId int) {
	fm.correct[nodeId] = false
	for _, f := range fm.failureCallback {
		f(nodeId)
	}
}

func (fm *failureManager) Subscribe(callback func(int)) {
	fm.failureCallback = append(fm.failureCallback, callback)
}
