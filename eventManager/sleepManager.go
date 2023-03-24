package eventManager

import (
	"fmt"
	"gomc/event"
	"gomc/scheduler"
	"runtime"
	"time"
)

type SleepManager struct {
	sch     scheduler.RunScheduler
	nextEvt chan error
}

func NewSleepManager(sch scheduler.RunScheduler, nextEvent chan error) *SleepManager {
	return &SleepManager{
		sch:     sch,
		nextEvt: nextEvent,
	}
}

func (sm *SleepManager) SleepFunc(id int) func(time.Duration) {
	return func(_ time.Duration) {
		sleepChan := make(chan time.Time)
		_, file, line, _ := runtime.Caller(1)
		evt := event.NewSleepEvent(fmt.Sprintf("File: %v, Line: %v", file, line), id, sleepChan)
		sm.sch.AddEvent(evt)
		// Inform the simulator that the process is currently waiting for a scheduled timeout
		// The simulator can now proceed with scheduling events
		sm.nextEvt <- nil

		// Wait until the event is executed before returning
		<-sleepChan
	}
}
