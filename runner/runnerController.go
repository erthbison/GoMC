package runner

import (
	"fmt"
	"gomc/event"
	"reflect"
)

// Controls the running of the algorithms.
//
// Manages the nodes in the system and forwards command to the specified nodes.
// Aggregates records from the nodes and sends them to subscribed users.
type RunnerController[T, S any] struct {
	// Buffer size of the record channels
	recordChanBuffer int

	// channel used to send subscribe channels from users
	subscribeRecordChan chan chan Record

	// map of all the node controllers in the system
	nodes map[int]*nodeController[T, S]

	// Callback functions that have subscribed to node crash updates
	crashSubscribes map[int]func(id int, status bool)

	// Id of the next request
	requestId int

	//
	stop chan bool
}

// Create a new EventController
//
// recordChanBuffer specifies the buffer size of the channel used to receive records from the nodes.
func NewEventController[T, S any](recordChanBuffer int) *RunnerController[T, S] {
	// The event controller is initially stopped and is started when the main function is called,
	stop := make(chan bool)
	close(stop)
	return &RunnerController[T, S]{
		subscribeRecordChan: make(chan chan Record),
		recordChanBuffer:    recordChanBuffer,

		crashSubscribes: make(map[int]func(id int, status bool)),

		stop: stop,
	}
}

// Subscribe to the records that are created by the Runner
func (ec *RunnerController[T, S]) Subscribe() <-chan Record {
	outRecordChan := make(chan Record, ec.recordChanBuffer)
	ec.subscribeRecordChan <- outRecordChan
	return outRecordChan
}

// Add the specified event to the correct node.
//
// Must be called after the Main loop has been started.
func (ec *RunnerController[T, S]) AddEvent(evt event.Event) {
	if ec.isClosed() {
		return
	}
	id := evt.Target()
	node, ok := ec.nodes[id]
	if !ok {
		return
	}
	node.addEvent(evt)
}

// Send the status of the previous event to the specified node.
//
// The status is provided as the err parameter and the node is specified using the id.
// Must be called after the Main loop has been started.
func (ec *RunnerController[T, S]) NextEvent(err error, id int) {
	if ec.isClosed() {
		return
	}
	node, ok := ec.nodes[id]
	if !ok {
		return
	}
	node.nextEvent(err)
}

// The Main loop of the Runner Controller.
//
// Starts the record loop that collects and forwards record.
// Initializes the nodes and starts their separate main loops.
func (ec *RunnerController[T, S]) MainLoop(nodes map[int]*T, eventChanBuffer int, crashFunc func(*T), getState func(*T) S) {
	ec.stop = make(chan bool)

	// Receive the records of events and states from the different nodes
	inRecordChan := make(chan Record, ec.recordChanBuffer)
	// Start the record loop that collects and forwards records
	go ec.recordLoop(inRecordChan)

	// Create the nodeManagers and start their main loop
	ec.nodes = make(map[int]*nodeController[T, S])
	for id, node := range nodes {
		nc := NewNodeController(id, node, getState, crashFunc, inRecordChan, eventChanBuffer)
		ec.nodes[id] = nc
		go nc.Main()
	}
}

// Collects and forwards records.
func (ec *RunnerController[T, S]) recordLoop(inRecordChan <-chan Record) {
	outRecordChan := make([]chan<- Record, 0)
	for {
		select {
		case rec := <-inRecordChan:
			// Forwards the received record to subscribed users
			for _, c := range outRecordChan {
				c <- rec
			}
		case c := <-ec.subscribeRecordChan:
			// Add the received channel to the subscribed channels
			outRecordChan = append(outRecordChan, c)
		case <-ec.stop:
			// Close all out channels and stop the loop
			for _, c := range outRecordChan {
				close(c)
			}
			return
		}
	}
}

// Stop the running of the algorithm
//
// Stop the execution of events on the nodes.
// Stop handling new records and close the record loops
//
// Must be called after the main loop has been started.
func (ec *RunnerController[T, S]) Stop() {
	if ec.isClosed() {
		return
	}
	for _, n := range ec.nodes {
		n.Close()
	}
	// Signal to the recordLoop to close all record channels
	close(ec.stop)

	// Remove the nodes.
	ec.nodes = nil
}

// Pause the execution of events on the specified node.
//
// Must be called after the main loop has been started.
func (ec *RunnerController[T, S]) Pause(id int) error {
	if ec.isClosed() {
		return fmt.Errorf("RunnerController: Runner Controller has been stopped. No more commands can be executed.")
	}
	node, ok := ec.nodes[id]
	if !ok {
		return fmt.Errorf("RunnerController: No node with the provided id. Provided id: %v", id)
	}
	node.Pause()
	return nil
}

// Resume the execution of events on the specified node.
//
// Must be called after the main loop has been started.
func (ec *RunnerController[T, S]) Resume(id int) error {
	if ec.isClosed() {
		return fmt.Errorf("RunnerController: Runner Controller has been stopped. No more commands can be executed.")
	}
	node, ok := ec.nodes[id]
	if !ok {
		return fmt.Errorf("RunnerController: No node with the provided id. Provided id: %v", id)
	}
	node.Resume()
	return nil
}

// Crash the specified node.
//
// Must be called after the main loop has been started.
func (ec *RunnerController[T, S]) CrashNode(id int) error {
	if ec.isClosed() {
		return fmt.Errorf("RunnerController: Runner Controller has been stopped. No more commands can be executed.")
	}
	node, ok := ec.nodes[id]
	if !ok {
		return fmt.Errorf("RunnerController: No node with the provided id. Provided id: %v", id)
	}
	node.Close()

	ec.sendCrashNotification(id)
	return nil
}

// Send crash notification to all subscribed nodes
func (ec *RunnerController[T, S]) sendCrashNotification(crashedId int) {
	for id, f := range ec.crashSubscribes {
		node, ok := ec.nodes[id]
		if !ok {
			continue
		}
		node.addEvent(event.NewCrashDetection(id, crashedId, f))
	}
}

// Subscribe to notifications about changes in the status of the nodes.
//
// id is the id of the node that subscribes to status changes.
// The provided callback function is called When the status of a node changes.
func (ec *RunnerController[T, S]) CrashSubscribe(id int, callback func(id int, status bool)) {
	ec.crashSubscribes[id] = callback
}

// Send a new request to the specified node.
//
// id is the id of the node that will receive the request.
// method is a string with the name of the method that will be called on the node.
// params is the parameters that will be passed to the method.
//
// Must be called after the main loop has been started.
func (ec *RunnerController[T, S]) NewRequest(id int, method string, params []reflect.Value) error {
	if ec.isClosed() {
		return fmt.Errorf("RunnerController: Runner Controller has been stopped. No more commands can be executed.")
	}

	node, ok := ec.nodes[id]
	if !ok {
		return fmt.Errorf("RunnerController: No node with the provided id. Provided id: %v", id)
	}

	node.addEvent(event.NewFunctionEvent(ec.requestId, id, method, params...))
	ec.requestId++
	return nil
}

// Returns true if the running has been stopped.
// false otherwise.
func (ec *RunnerController[T, S]) isClosed() bool {
	select {
	case <-ec.stop:
		return true
	default:
		return false
	}
}
