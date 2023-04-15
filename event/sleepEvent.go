package event

import (
	"fmt"
	"hash/maphash"
	"time"
)

type SleepEvent struct {
	// An event representing a timeout.
	// It is analogous to time.Sleep and can be used to represent timeouts for example for a failure detector
	caller      string
	target      int // The id of the target node
	timeoutChan chan time.Time

	id uint64
}

func NewSleepEvent(caller string, target int, timeoutChan chan time.Time) SleepEvent {
	evt := SleepEvent{
		caller:      caller,
		target:      target,
		timeoutChan: timeoutChan,

		id: maphash.String(EventHash, fmt.Sprint("Sleep ", target, caller)),
	}
	return evt
}

func (se SleepEvent) Id() uint64 {
	return se.id
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
