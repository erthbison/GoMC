package runner

import (
	"fmt"
	"gomc/event"
	"reflect"
)

type RunnerController[T, S any] struct {
	// Receive the records of events and states from the different nodes
	inRecordChan        chan Record
	outRecordChan       []chan<- Record
	subscribeRecordChan chan chan Record

	// map of all the node controllers in the system
	nodes map[int]*nodeController[T, S]

	// Callback functions that have subscribed to node crash updates
	crashSubscribes []func(id int, status bool)

	requestId int

	stop chan bool
}

func NewEventController[T, S any](recordChanBuffer int) *RunnerController[T, S] {
	return &RunnerController[T, S]{
		inRecordChan:  make(chan Record, recordChanBuffer),
		outRecordChan: make([]chan<- Record, 0),

		subscribeRecordChan: make(chan chan Record),

		crashSubscribes: make([]func(id int, status bool), 0),

		stop: make(chan bool),
	}
}

// Subscribe to a copy of the records that are sent by the runner
func (ec *RunnerController[T, S]) Subscribe() <-chan Record {
	outRecordChan := make(chan Record)
	ec.subscribeRecordChan <- outRecordChan
	return outRecordChan
}

func (ec *RunnerController[T, S]) AddEvent(evt event.Event) {
	id := evt.Target()
	node, ok := ec.nodes[id]
	if !ok {
		return
	}
	node.addEvent(evt)
}

func (ec *RunnerController[T, S]) NextEvent(err error, id int) {
	node, ok := ec.nodes[id]
	if !ok {
		return
	}
	node.nextEvent(err)
}

func (ec *RunnerController[T, S]) MainLoop(nodes map[int]*T, eventChanBuffer int, crashFunc func(*T), getState func(*T) S) {
	go func(inRecordChan <-chan Record) {
		for {
			select {
			case rec := <-inRecordChan:
				for _, c := range ec.outRecordChan {
					c <- rec
				}
			case c := <-ec.subscribeRecordChan:
				ec.outRecordChan = append(ec.outRecordChan, c)
			case <-ec.stop:
				return
			}
		}
	}(ec.inRecordChan)
	ec.nodes = make(map[int]*nodeController[T, S])
	for id, node := range nodes {
		nc := NewNodeController(id, node, getState, crashFunc, ec.inRecordChan, eventChanBuffer)
		ec.nodes[id] = nc
		go nc.Main()
	}
}

func (ec *RunnerController[T, S]) Stop() {
	for _, n := range ec.nodes {
		n.Close()
	}
	close(ec.stop)
	for _, c := range ec.outRecordChan {
		close(c)
	}
}

func (ec *RunnerController[T, S]) Pause(id int) error {
	node, ok := ec.nodes[id]
	if !ok {
		return fmt.Errorf("RunnerController: No node with the provided id. Provided id: %v", id)
	}
	node.Pause()
	return nil
}

func (ec *RunnerController[T, S]) Resume(id int) error {
	node, ok := ec.nodes[id]
	if !ok {
		return fmt.Errorf("RunnerController: No node with the provided id. Provided id: %v", id)
	}
	node.Resume()
	return nil
}

func (ec *RunnerController[T, S]) CrashNode(id int) error {
	node, ok := ec.nodes[id]
	if !ok {
		return fmt.Errorf("RunnerController: No node with the provided id. Provided id: %v", id)
	}
	node.Close()
	for _, f := range ec.crashSubscribes {
		f(id, false)
	}
	return nil
}

func (ec *RunnerController[T, S]) CrashSubscribe(f func(id int, status bool)) {
	ec.crashSubscribes = append(ec.crashSubscribes, f)
}

func (ec *RunnerController[T, S]) NewRequest(id int, method string, params []reflect.Value) error {
	ec.AddEvent(event.NewFunctionEvent(ec.requestId, id, method, params...))
	ec.requestId++
	return nil
}
