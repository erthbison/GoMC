package state

import (
	"fmt"
	"gomc/event"
)

// A Record of an event
//
// Stores the id and the string representation of the event
type EventRecord struct {
	Id   event.EventId
	Repr string
}

func (er EventRecord) String() string {
	return er.Repr
}

// Create a EventRecord from an event.
//
// If the event is nil, create a EventRecord with zero value for all fields.
func CreateEventRecord(evt event.Event) EventRecord {
	if evt != nil {
		return EventRecord{
			Id:   evt.Id(),
			Repr: fmt.Sprint(evt),
		}
	}
	return EventRecord{
		Id:   "",
		Repr: "",
	}
}
