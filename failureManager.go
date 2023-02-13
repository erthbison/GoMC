package gomc

import "gomc/scheduler"

type failureManager[T any] struct {
	sch scheduler.Scheduler[T]

	correct         map[int]bool
	failureCallback []func(int)
}

func NewFailureManager[T any](sch scheduler.Scheduler[T]) *failureManager[T] {
	return &failureManager[T]{
		correct:         make(map[int]bool),
		sch:             sch,
		failureCallback: make([]func(int), 0),
	}
}

func (fm *failureManager[T]) CorrectNodes() map[int]bool {
	return fm.correct
}

func (fm *failureManager[T]) NodeCrash(nodeId int) {
	fm.sch.NodeCrash(nodeId)
	fm.correct[nodeId] = false
	for _, f := range fm.failureCallback {
		f(nodeId)
	}
}

func (fm *failureManager[T]) Subscribe(callback func(int)) {
	fm.failureCallback = append(fm.failureCallback, callback)
}
