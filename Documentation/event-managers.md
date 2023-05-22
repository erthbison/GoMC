# Event Managers

Event Managers are the hooks that are inserted into the algorithm.
They allow Go-MC to record and control the execution of events. 

To use Go-MC with new frameworks it might be useful to implement new Event Managers. 
When implementing a new Event Manager it is also common to implement a new `Event`.
There are no interface that must be implemented by the Event Managers, but Go-MC makes some assumptions about the `Event` and Event Managers that must hold.

## Events

An `Event` is assumed to be atomic and deterministic. 
Go-MC uses the `NextEvent` signal to ensure that an `Event` remains atomic. 
A new `Event` will not be scheduled before the previous `Event` sends a `NextEvent` signal.
It is important that the node has completed all executions when the `NextEvent` signal is sent, otherwise the atomicity of events might be broken.

Some schedulers rely on the determinism of an `Event` to be able to build a view of the state space and properly perform scheduling. 
We therefore expect an `Event` to be deterministic.
This means that the event should not contain randomness that is not controlled by Go-MC.

## Event Managers

The Event Manager is usually instantiated with a variable of the type `SimulationParameters`.
The type contains the specific simulation parameters used in this run. 
Specifically, it contains `NextEvt`, `CrashSubscribe` and `EventAdder`. 
`NextEvt` is a function used to send the `NextEvent` signal to Go-MC.
`CrashSubscribe` is a function used by nodes to subscribe to status changes. 
`EventAdder` is a type implementing the `EventAdder` interface. 
When simulating the type is a Scheduler and when running the algorithm it is a `RunnerController`.
New events should be added to the `EventAdder` when detected.

It is important that the Event Manager is instantiated with the `SimulationParameters` for the current run.
The correct `SimulationParameters` for a run are provided when instantiating the nodes for the run. 

```go
// Stores the SimulationParameters used in the specific run of the simulation
//
// The parameters are used to initialize the EventManager and to subscribe to the failure detector
type SimulationParameters struct {
	// Signal the status of the executed event to the main loop.
	NextEvt func(error, int)

	// Used by node to subscribe to status updates from the failure detector
	CrashSubscribe func(NodeId int, callback func(id int, status bool))

	// Add events to the EventAdder used by the simulation
	EventAdder EventAdder
}
```
