package state

import "fmt"

// The global state of the nodes at one time slot of the simulation
type GlobalState[S any] struct {
	// A map storing the local state of the nodes. The map stores (id, state) combination.
	LocalStates map[int]S
	// A map storing the status of the node. The map stores (id, status) combination.
	// If status is true, the node with id "id" is correct, otherwise the node is true.
	// All nodes are represented in the map.
	Correct map[int]bool
	// A record of the event that caused the transition into this state
	Evt EventRecord
}

func (gs GlobalState[S]) String() string {
	crashed := []int{}
	for id, status := range gs.Correct {
		if !status {
			crashed = append(crashed, id)
		}
	}
	return fmt.Sprintf("Evt: %v\t States: %v\t Crashed: %v\t", gs.Evt, gs.LocalStates, crashed)
}
