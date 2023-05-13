package eventManager

import "gomc/event"

// A type that can receive Events
type EventAdder interface {
	// Add the event to the EventAdder
	AddEvent(event.Event)
}
