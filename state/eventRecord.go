package state

import (
	"fmt"
	"gomc/event"
)

// A type providing a record of some event
type EventRecord struct {
	Id   string
	Repr string
}

func (er EventRecord) String() string {
	return er.Repr
}

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
