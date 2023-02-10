package gomc

import (
	"fmt"
	"gomc/event"
	"gomc/scheduler"
	"runtime"
	"time"
)

type SleepManager[T any] struct {
	sch        scheduler.Scheduler[T]
	nextEvt    chan error
	sleepChans map[string]chan time.Time
}

func NewSleepManager[T any](sch scheduler.Scheduler[T], nextEvent chan error) *SleepManager[T] {
	return &SleepManager[T]{
		sch:        sch,
		nextEvt:    nextEvent,
		sleepChans: make(map[string]chan time.Time),
	}
}

func (sm *SleepManager[T]) SleepFunc(id int) func(time.Duration) {
	return func(_ time.Duration) {
		_, file, line, _ := runtime.Caller(1)
		evt := event.NewSleepEvent[T](fmt.Sprintf("File: %v, Line: %v", file, line), id, sm.sleepChans)
		sm.sch.AddEvent(evt)
		// Inform the simulator that the process is currently waiting for a scheduled timeout
		// The simulator can now proceed with scheduling events
		sm.nextEvt <- nil

		// Wait until the event is executed before returning
		<-sm.sleepChans[evt.Id()]
	}
}
