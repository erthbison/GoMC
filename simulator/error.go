package simulator

import "fmt"

// Aggregates the errors that ocurred during a simulation
type simulationError struct {
	errorSlice []error
}

func (se simulationError) Error() string {
	return fmt.Sprintf("Simulator: %v Errors occurred running simulations. \nError 1: %v", len(se.errorSlice), se.errorSlice[0])
}
