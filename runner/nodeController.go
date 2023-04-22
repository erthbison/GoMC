package runner

import (
	"gomc/event"
)

type nodeController[T, S any] struct {
	id      int
	node    *T
	crashed bool

	getState  func(*T) S
	crashFunc func(*T)

	recordChan chan Record

	pauseChan  chan bool
	resumeChan chan bool

	eventQueue  chan event.Event
	nextEvtChan chan error
}

func NewNodeController[T, S any](id int, node *T, getState func(*T) S, crashFunc func(*T), evtChan chan Record) *nodeController[T, S] {
	return &nodeController[T, S]{
		id:   id,
		node: node,

		recordChan: evtChan,

		crashFunc: crashFunc,
		getState:  getState,

		pauseChan:  make(chan bool, 1),
		resumeChan: make(chan bool, 1),

		eventQueue:  make(chan event.Event, 1000),
		nextEvtChan: make(chan error),
	}
}

func (nc *nodeController[T, S]) Main() {
	for {
		select {
		case <-nc.pauseChan:
			// If there is a pause message wait for the next resume message
			<-nc.resumeChan
		case evt, ok := <-nc.eventQueue:
			if !ok {
				return
			}
			nc.recordEvent(evt, true)
			go evt.Execute(nc.node, nc.nextEvtChan)
			<-nc.nextEvtChan
			nc.recordState()
		}
	}
}

func (nc *nodeController[T, S]) recordState() {
	state := nc.getState(nc.node)
	nc.recordChan <- StateRecord{
		target: nc.id,
		state:  state,
	}
}

func (nc *nodeController[T, S]) recordEvent(evt event.Event, isExecuting bool) {
	if msg, ok := evt.(event.MessageEvent); ok {
		nc.recordChan <- MessageRecord{
			from: msg.From(),
			to:   msg.To(),
			sent: !isExecuting,
			evt:  msg,
		}
		return
	}

	// If we are not executing the event here and it is not a message event then we do not record it
	// We will record it later when it is actually executed
	if !isExecuting {
		return
	}

	// Want to add record non-message events as well
	nc.recordChan <- ExecutionRecord{
		target: evt.Target(),
		evt:    evt,
	}
}

func (nc *nodeController[T, S]) addEvent(evt event.Event) {
	if nc.crashed {
		return
	}

	nc.recordEvent(evt, false)
	nc.eventQueue <- evt
}

func (nc *nodeController[T, S]) nextEvent(err error) {
	nc.nextEvtChan <- err
}

func (nc *nodeController[T, S]) Pause() {
	nc.pauseChan <- true
}

func (nc *nodeController[T, S]) Resume() {
	nc.resumeChan <- true
}

func (nc *nodeController[T, S]) Close() {
	close(nc.eventQueue)
}

func (nc *nodeController[T, S]) Crash() {
	nc.crashed = true
}
