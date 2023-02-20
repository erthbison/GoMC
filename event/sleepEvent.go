package event

import (
	"fmt"
	"time"
)

type SleepEvent struct {
	// An event representing a timeout.
	// It is analogous to time.Sleep and can be used to represent timeouts for example for a failure detector
	caller      string
	target      int // The id of the target node
	timeoutChan chan time.Time
}

func NewSleepEvent(caller string, target int, timeoutChan chan time.Time) SleepEvent {
	evt := SleepEvent{
		caller:      caller,
		target:      target,
		timeoutChan: timeoutChan,
	}
	return evt
}

func (se SleepEvent) Id() string {
	return fmt.Sprintf("Sleep Target: %v Caller: %v", se.target, se.caller)
}

func (se SleepEvent) String() string {
	return fmt.Sprintf("{Sleep Target: %v}", se.target)
}

func (se SleepEvent) Execute(node any, _ chan error) {
	// Send a signal on the timeout channel
	// Don't signal on the error channel since the event that was paused by the sleep event will continue running and the simulator can therefore not begin collecting state yet.
	// The event that continues after the sleep event will signal to the simulator when it has completed.
	se.timeoutChan <- time.Time{}
}

func (se SleepEvent) Target() int {
	return se.target
}
