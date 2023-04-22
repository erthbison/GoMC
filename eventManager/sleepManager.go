package eventManager

import (
	"fmt"
	"gomc/event"
	"runtime"
	"time"
)

type SleepManager struct {
	ea      EventAdder
	nextEvt func(error, int)
}

func NewSleepManager(ea EventAdder, nextEvent func(error, int)) *SleepManager {
	return &SleepManager{
		ea:      ea,
		nextEvt: nextEvent,
	}
}

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
