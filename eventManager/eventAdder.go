package eventManager

import "gomc/event"

type EventAdder interface {
	AddEvent(event.Event)
}
