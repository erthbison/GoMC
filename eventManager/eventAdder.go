package eventManager

import "gomc/event"

// A type that can receive Events
type EventAdder interface {
	// Add the event to the EventAdder
	// It must be safe to add events from different goroutines.
	AddEvent(event.Event)
}
