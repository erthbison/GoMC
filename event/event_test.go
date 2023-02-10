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
	tmp := map[int]*node{}
	chanMap := make(map[string]chan time.Time)
	foo := NewSleepEvent[node]("Foo", chanMap)
	go func() {
		<-chanMap[foo.Id()]
	}()
	errChan := make(chan error)
	foo.Execute(tmp, errChan)
}

func TestSleepEventOnSameLocation(t *testing.T) {
	// Test that two Sleep events with the same id are interchangeable
	tmp := map[int]*node{}
	chanMap := make(map[string]chan time.Time)
	foo1 := NewSleepEvent[node]("Foo", chanMap)
	foo2 := NewSleepEvent[node]("Foo", chanMap)
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
	foo1.Execute(tmp, errChan)
	foo2.Execute(tmp, errChan)
	wg.Wait()
}
