package eventManager

import "gomc/event"

// A type that can receive Events
type EventAdder interface {
	// Add an event
	AddEvent(event.Event)
}
