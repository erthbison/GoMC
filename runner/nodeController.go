package runner

import (
	"gomc/event"
)

// The nodeController is used to control the execution of events on a single node
//
// It manages the events that should be executed on a node, and ensures that they are executed in a sequential order.
//
// It also executes the commands on the node.
// Commands should be given from the same goroutine.
//
// Finally it records the execution of events and the changes of states of the node and sends it on the recordChan.
type nodeController[T, S any] struct {
	id   int
	node *T

	// Is closed if the node has crashed. Otherwise it is open.
	// No messages are sent on the channel.
	crashed chan bool

	// Collects the state from the node
	getState func(*T) S
	// Crash/stop the node
	crashFunc func(*T)

	// Send updates of events and states to the user
	recordChan chan Record

	// Used to pause execution of events
	pauseChan chan bool
	// Used to pause and resume execution of events
	resumeChan chan bool
	// Closed if the execution of events on the node is paused.
	paused chan bool

	// Pending events for the node
	eventQueue chan event.Event
	// Signal when to begin the next event for the node
	nextEvtChan chan error
}

// Create a new NodeController
//
// Specify the id of the node and the instance of the node.
// Also requires a function collecting the state of the node and performing a crash on the node.
// The provided recordChan is used to send records on.
// eventQueueBuffer specifies the number of messages that can be stored at a time.
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

// Start the main loop of the nodeController
//
// The main loop executes events in a sequential manner.
// It sends records of the event that is executed and the state after the event is executed.
//
// The main loop can be paused by calling the Pause method and resumed by calling the resume method.
func (nc *nodeController[T, S]) Main() {
	// When the main loop stops: close the node
	defer nc.crashFunc(nc.node)
	for {
		select {
		case <-nc.pauseChan:
			// Pause the execution of events
			if _, ok := <-nc.resumeChan; !ok {
				return
			}
		case evt, ok := <-nc.eventQueue:
			// Execute events
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

// Record the current state of the node
func (nc *nodeController[T, S]) recordState() {
	nc.recordChan <- StateRecord[S]{
		target: nc.id,
		State:  nc.getState(nc.node),
	}
}

// Record the provided event
// isExecuting is true if the event is executed now, false if it is added to the nodeController.
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

// Add the event to the nodeController
func (nc *nodeController[T, S]) addEvent(evt event.Event) {
	if nc.isCrashed() {
		return
	}

	nc.recordEvent(evt, false)
	nc.eventQueue <- evt
}

// Signal to the main loop that the event has been completed.
// It can now collect state and proceed to execute the next event.
func (nc *nodeController[T, S]) nextEvent(err error) {
	if nc.isCrashed() {
		return
	}
	nc.nextEvtChan <- err
}

// Pause the execution of events on the node.
func (nc *nodeController[T, S]) Pause() {
	if nc.isCrashed() {
		return
	}
	if nc.isPaused() {
		return
	}
	close(nc.paused)
	nc.pauseChan <- true
}

// Resume the execution of events on the node.
func (nc *nodeController[T, S]) Resume() {
	if nc.isCrashed() {
		return
	}
	if !nc.isPaused() {
		return
	}
	nc.paused = make(chan bool)
	nc.resumeChan <- true
}

// Stop the node.
// Stops the execution of events on the main loop and calls the crashFunc on the node.
func (nc *nodeController[T, S]) Close() {
	if nc.isCrashed() {
		return
	}
	close(nc.crashed)
	close(nc.eventQueue)
	close(nc.resumeChan)
}

// Returns true if the node has crashed. False otherwise.
func (nc *nodeController[T, S]) isCrashed() bool {
	select {
	case <-nc.crashed:
		return true
	default:
		return false
	}
}

// Returns true if the execution of events is paused. False otherwise.
func (nc *nodeController[T, S]) isPaused() bool {
	select {
	case <-nc.paused:
		return true
	default:
		return false
	}
}
