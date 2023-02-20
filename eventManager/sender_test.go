package eventManager

import (
	"gomc/event"
	"testing"
)

func TestSender(t *testing.T) {
	// Basic test testing that it the sender can send a message.
	sch := NewMockScheduler()
	sender := NewSender(sch)
	send := sender.SendFunc(0)
	send(0, "Foo", []byte("Foo"))
	out, _ := sch.GetEvent()
	expected := event.NewMessageEvent(0, 0, "Foo", []byte("Foo"))
	if out.Id() != expected.Id() {
		t.Fatalf("Unexpected event. Got: %v. Expected: %v", out, expected)
	}
}
