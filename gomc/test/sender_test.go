package gomc_test

import (
	"experimentation/gomc"
	"testing"
)

func TestSender(t *testing.T) {
	// Basic test testing that it the sender can send a message.
	sch := NewMockScheduler()
	sender := gomc.NewSender[node](sch)

	go sender.Send(0, 0, "Foo", []byte("Foo"))
	out := <-sch.inEvent
	expected := gomc.NewMessageEvent[node](0, 0, "Foo", []byte("Foo"))
	if out.Id() != expected.Id() {
		t.Fatalf("Unexpected event. Got: %v. Expected: %v", out, expected)
	}
}
