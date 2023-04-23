package event

import "fmt"

type CrashDetection struct {
	targetId    int
	crashedNode int

	callback func(id int, status bool)

	evtId EventId
}

func NewCrashDetection(targetNode int, crashedNode int, callback func(int, bool)) CrashDetection {
	return CrashDetection{
		targetId:    targetNode,
		crashedNode: crashedNode,

		callback: callback,

		evtId: EventId(fmt.Sprint("CrashDetection", targetNode, crashedNode)),
	}
}

func (cd CrashDetection) String() string {
	return fmt.Sprintf("{CrashDetection Target: %v. Crashed Node: %v}", cd.targetId, cd.crashedNode)
}

// An id that identifies the event.
// Two events that provided the same input state results in the same output state should have the same id
//
// New event implementations should include a identifier of the event type to prevent accidental collisions with other implementations
func (cd CrashDetection) Id() EventId {
	return cd.evtId
}

// A method executing the event.
// The event will be executed on a separate goroutine.
// It should signal on the channel if it is clear for the simulator to proceed to processing of the state and the next event.
// Panics raised while executing the event is recovered by the simulator and returned as errors
func (cd CrashDetection) Execute(node any, errorChan chan error) {
	cd.callback(cd.crashedNode, false)
	errorChan <- nil
}

// The id of the target node, i.e. the node whose state will be changed by the event executing.
// Is used to identify if an event is still enabled, or if it has been disabled, e.g. because the node crashed.
func (cd CrashDetection) Target() int {
	return cd.targetId
}
