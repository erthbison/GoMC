package gomc

import (
	"fmt"
	"runtime"
	"time"
)

type SleepManager[T any, S any] struct {
	sim        *Simulator[T, S]
	sch        Scheduler[T]
	sleepChans map[string]chan time.Time
}

func NewSleepManager[T any, S any](sch Scheduler[T], sim *Simulator[T, S]) *SleepManager[T, S] {
	return &SleepManager[T, S]{
		sim:        sim,
		sch:        sch,
		sleepChans: make(map[string]chan time.Time),
	}
}

func (sm *SleepManager[T, S]) Sleep(_ time.Duration) {
	_, file, line, _ := runtime.Caller(1)
	evt := NewSleepEvent[T](fmt.Sprintf("File: %v, Line: %v", file, line), sm.sleepChans)
	sm.sch.AddEvent(evt)
	// Inform the simulator that the process is currently waiting for a scheduled timeout
	// The simulator can now proceed with scheduling events
	sm.sim.nextEvt <- nil
	<-sm.sleepChans[evt.Id()]
}
