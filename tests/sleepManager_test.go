package gomc_test

import (
	"gomc"
	"testing"
	"time"
)

func TestSleepManager(t *testing.T) {
	sch := NewMockScheduler()
	nextEvent := make(chan error)
	sm := gomc.NewSleepManager[node](sch, nextEvent)
	sleepFunc := sm.SleepFunc(0)
	notBlockedChan := make(chan bool)
	go func() {
		// Need to check that sleep actually blocks until the event is executed
		sleepFunc(time.Second)
		notBlockedChan <- true
	}()

	evt := <-sch.inEvent

	// Wait until the nextEvent signal is received than execute the event.
	// Like the simulator would
	select {
	case <-nextEvent:
		evt.Execute(map[int]*node{}, nextEvent)
	case <-notBlockedChan:
		t.Fatalf("Sleep returned before the event was executed")
	}
	select {
	case <-notBlockedChan:
	case <-time.After(time.Second * 5):
		t.Fatalf("Sleep is still blocking after executing the event")
	}

}
