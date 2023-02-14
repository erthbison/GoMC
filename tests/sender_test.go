package gomc_test

import (
	"gomc"
	"gomc/event"
	"testing"
)

func TestSender(t *testing.T) {
	// Basic test testing that it the sender can send a message.
	sch := NewMockScheduler()
	sender := gomc.NewSender(sch)
	send := sender.SendFunc(0)
	go send(0, "Foo", []byte("Foo"))
	out := <-sch.inEvent
	expected := event.NewMessageEvent(0, 0, "Foo", []byte("Foo"))
	if out.Id() != expected.Id() {
		t.Fatalf("Unexpected event. Got: %v. Expected: %v", out, expected)
	}
}
