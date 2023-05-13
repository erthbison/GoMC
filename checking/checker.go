package checking

import (
	"gomc/event"
	"gomc/state"
)

// The Checker verifies that properties hold for the state space.
type Checker[S any] interface {
	// Verify that the configured properties hold for the provided state space
	Check(root state.StateSpace[S]) CheckerResponse
}

// CheckerResponse is a response returned by a Checker
//
// Contains the result of checking the system.
type CheckerResponse interface {
	// Create a response.
	//
	// Returns a boolean that is true if all properties hold, false otherwise.
	// Returns a string describing the response.
	// This should include a detailed description of which property is violated and the run which caused it to be violated.
	Response() (bool, string)

	// Export the run which caused a property to be violated
	//
	// If a property was violated it will return a slice containing the sequence of EventIds of the events in the run.
	// Otherwise it will return an empty slice.
	Export() []event.EventId
}
