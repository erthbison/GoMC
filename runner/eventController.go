package runner

import (
	"gomc/event"
)

type EventController[T, S any] struct {
	msgChan chan Record

	nodes map[int]*nodeController[T, S]
}

func NewEventController[T, S any]() *EventController[T, S] {
	return &EventController[T, S]{
		msgChan: make(chan Record, 1000),
	}
}

func (ec *EventController[T, S]) Subscribe() <-chan Record {
	return ec.msgChan
}

func (ec *EventController[T, S]) AddEvent(evt event.Event) {
	ec.nodes[evt.Target()].addEvent(evt)
}

func (ec *EventController[T, S]) NextEvent(err error, id int) {
	ec.nodes[id].nextEvent(err)
}

func (ec *EventController[T, S]) MainLoop(nodes map[int]*T, crashFunc func(*T), getState func(*T) S) {
	ec.nodes = make(map[int]*nodeController[T, S])
	for id, node := range nodes {
		nc := NewNodeController(id, node, getState, crashFunc, ec.msgChan)
		ec.nodes[id] = nc
		go nc.Main()
	}
}

func (ec *EventController[T, S]) Stop() {
	for _, n := range ec.nodes {
		n.Close()
	}
	close(ec.msgChan)
}

func (ec *EventController[T, S]) Pause(id int) error {
	ec.nodes[id].Pause()
	return nil
}

func (ec *EventController[T, S]) Resume(id int) error {
	ec.nodes[id].Resume()
	return nil
}

func (ec *EventController[T, S]) CrashNode(id int) {
	ec.nodes[id].Crash()
}
