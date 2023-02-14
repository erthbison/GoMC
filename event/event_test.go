package event

import (
	"sync"
	"testing"
	"time"
)

type node struct{}

func (n *node) Foo(from, to int, msg []byte) {}
func (n *node) Bar(from, to int, msg []byte) {}

func TestSleepEvent(t *testing.T) {
	chanMap := make(map[string]chan time.Time)
	foo := NewSleepEvent("Foo", 0, chanMap)
	go func() {
		<-chanMap[foo.Id()]
	}()
	errChan := make(chan error)
	foo.Execute(&node{}, errChan)
}

func TestSleepEventOnSameLocation(t *testing.T) {
	// Test that two Sleep events with the same id are interchangeable
	chanMap := make(map[string]chan time.Time)
	foo1 := NewSleepEvent("Foo", 0, chanMap)
	foo2 := NewSleepEvent("Foo", 0, chanMap)
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		<-chanMap[foo1.Id()]
		wg.Done()
	}()
	go func() {
		<-chanMap[foo2.Id()]
		wg.Done()
	}()
	errChan := make(chan error)
	foo1.Execute(&node{}, errChan)
	foo2.Execute(&node{}, errChan)
	wg.Wait()
}
