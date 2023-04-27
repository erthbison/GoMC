package runner

import (
	"gomc/event"
)

type nodeController[T, S any] struct {
	id   int
	node *T

	// Is closed if the node has crashed. otherwise open.
	// No messages are sent on the channel
	crashed chan bool

	// Collects the state from the node
	getState func(*T) S
	// Crash/stop the node
	crashFunc func(*T)

	// Send updates of events and states to the user
	recordChan chan Record

	// Used to pause and resume execution of events
	pauseChan  chan bool
	resumeChan chan bool
	paused     chan bool

	// Pending events for the node
	eventQueue chan event.Event
	// Signal when to begin the next event for the node
	nextEvtChan chan error
}

func NewNodeController[T, S any](id int, node *T, getState func(*T) S, crashFunc func(*T), recordChan chan Record, eventQueueBuffer int) *nodeController[T, S] {
	return &nodeController[T, S]{
		id:   id,
		node: node,

		crashed: make(chan bool),

		recordChan: recordChan,

		crashFunc: crashFunc,
		getState:  getState,

		pauseChan:  make(chan bool),
		resumeChan: make(chan bool),
		paused:     make(chan bool),

		eventQueue:  make(chan event.Event, eventQueueBuffer),
		nextEvtChan: make(chan error),
	}
}

func (nc *nodeController[T, S]) Main() {
	for {
		select {
		case <-nc.pauseChan:
			close(nc.paused)
			if _, ok := <-nc.resumeChan; !ok {
				return
			}
			nc.paused = make(chan bool)

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
	nc.recordChan <- StateRecord[S]{
		target: nc.id,
		State:  nc.getState(nc.node),
	}
}

func (nc *nodeController[T, S]) recordEvent(evt event.Event, isExecuting bool) {
	if msg, ok := evt.(event.MessageEvent); ok {
		nc.recordChan <- MessageRecord{
			From: msg.From(),
			To:   msg.To(),
			Sent: !isExecuting,
			Evt:  msg,
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
		Evt:    evt,
	}
}

func (nc *nodeController[T, S]) addEvent(evt event.Event) {
	if nc.isCrashed() {
		return
	}

	nc.recordEvent(evt, false)
	nc.eventQueue <- evt
}

func (nc *nodeController[T, S]) nextEvent(err error) {
	if nc.isCrashed() {
		return
	}
	nc.nextEvtChan <- err
}

func (nc *nodeController[T, S]) Pause() {
	if nc.isCrashed() {
		return
	}
	if nc.isPaused() {
		return
	}
	nc.pauseChan <- true
}

func (nc *nodeController[T, S]) Resume() {
	if nc.isCrashed() {
		return
	}
	if !nc.isPaused() {
		return
	}
	nc.resumeChan <- true
}

func (nc *nodeController[T, S]) Close() {
	if nc.isCrashed() {
		return
	}
	close(nc.crashed)

	nc.crashFunc(nc.node)
	close(nc.eventQueue)
	close(nc.resumeChan)
}

func (nc *nodeController[T, S]) isCrashed() bool {
	select {
	case <-nc.crashed:
		return true
	default:
		return false
	}
}

func (nc *nodeController[T, S]) isPaused() bool {
	select {
	case <-nc.paused:
		return true
	default:
		return false
	}
}
