package eventManager

import (
	"fmt"
	"gomc/event"
	"runtime"
	"time"
)

// An EventManager that is used to set timeouts
type SleepManager struct {
	ea      EventAdder
	nextEvt func(error, int)
}

// Create a SleepManager with the provided EventAdder and nextEvent function
func NewSleepManager(ea EventAdder, nextEvent func(error, int)) *SleepManager {
	return &SleepManager{
		ea:      ea,
		nextEvt: nextEvent,
	}
}

// Creates a sleep function that is used to create timeouts.
// The SleepFunc method is called with the id of the node that sets the timeout.
// Should create a new sleep function for each node.
//
// The returned sleep function imitates the signature of the time.Sleep function.
// The provided duration is ignored.
// It is assumed that at most one timeout is active at each node at teh same time.
func (sm *SleepManager) SleepFunc(id int) func(time.Duration) {
	return func(_ time.Duration) {
		sleepChan := make(chan time.Time)
		_, file, line, _ := runtime.Caller(1)
		evt := event.NewSleepEvent(fmt.Sprintf("File: %v, Line: %v", file, line), id, sleepChan)
		sm.ea.AddEvent(evt)
		// Inform the simulator that the process is currently waiting for a scheduled timeout
		// The simulator can now proceed with scheduling events
		sm.nextEvt(nil, id)

		// Wait until the event is executed before returning
		<-sleepChan
	}
}
