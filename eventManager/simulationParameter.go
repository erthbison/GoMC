package eventManager

// TODO: Update thesis to reflect the move of SimulationParameters into eventManager package

// Stores the SimulationParameters used in the specific run of the simulation
//
// The parameters are used by to initialize the EventManager and to subscribe to the failure detector
type SimulationParameters struct {
	// Signal the status of the executed event to the main loop.
	NextEvt func(error, int)

	// Used by node to subscribe to status updates from the failure detector
	CrashSubscribe func(NodeId int, callback func(id int, status bool))

	// Add events to the EventAdder used by the simulation
	EventAdder EventAdder
}
