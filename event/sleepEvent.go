package event

import (
	"fmt"
	"time"
)

type SleepEvent[T any] struct {
	// An event representing a timeout.
	// It is analogous to time.Sleep and can be used to represent timeouts for example for a failure detector
	caller      string
	target      int // The id of the target node
	timeoutChan map[string]chan time.Time
}

func NewSleepEvent[T any](caller string, target int, timeoutChan map[string]chan time.Time) SleepEvent[T] {
	waitChan := make(chan time.Time)
	evt := SleepEvent[T]{
		caller:      caller,
		target:      target,
		timeoutChan: timeoutChan,
	}
	if _, ok := timeoutChan[evt.Id()]; !ok {
		timeoutChan[evt.Id()] = waitChan
	}
	return evt
}

func (se SleepEvent[T]) Id() string {
	return fmt.Sprintf("Sleep Target: %v Caller: %v", se.target, se.caller)
}

func (se SleepEvent[T]) String() string {
	return fmt.Sprintf("{Sleep Target: %v Caller: %v}", se.target, se.caller)
}

func (se SleepEvent[T]) Execute(node map[int]*T, _ chan error) {
	// Send a signal on the timeout channel
	// Don't signal on the error channel since the event that was paused by the sleep event will continue running and the simulator can therefore not begin collecting state yet.
	// The event that continues after the sleep event will signal to the simulator when it has completed.
	se.timeoutChan[se.Id()] <- time.Time{}
}

func (se SleepEvent[T]) Target() int {
	return se.target
}
