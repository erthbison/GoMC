package config

import "gomc/scheduler"

// Configures the scheduler used by the simulation.
//
// The scheduler determines the method which will be used to explore the state space.
// It determines the order that events will be executed in.
// Default value is a Prefix Scheduler
type SchedulerOption struct {
	Sch scheduler.GlobalScheduler
}

func (so SchedulerOption) SimOpt() {}

// Configures the max depth of a run

// The depth of a run is the number of events that are executed in a run.
// If a simulation reaches the maxDepth it will stop.
// If a simulation is stopped early, we can not determine whether the algorithm is correct or not,
// we can only determine whether we found some violations in the recorded states.
// Decreasing the max depth will reduce the size of the state space.
// This will reduce memory and computation time requirements.
// Default value is 100
type MaxDepthOption struct{ MaxDepth int }

func (mdo MaxDepthOption) SimOpt() {}

// Configures the maximum number of runs that will be simulated.

// If the number of simulated runs reaches the maximum number of runs the simulation will stop.
// If a simulation is stopped early, we can not determine whether the algorithm is correct or not,
// we can only determine whether we found some violations in the recorded runs.
// Decreasing the maximum number of runs will reduce the size of the discovered state space.
// This will reduce memory and computation time requirements.
// Default value is 1000
type MaxRunsOption struct{ MaxRuns int }

func (mro MaxRunsOption) SimOpt() {}

// Configures the number of runs that are simulated concurrently.
//
// Higher values will reduce simulation time, but require more memory.
// Default value is GOMAXPROCS
type NumConcurrentOption struct{ N int }

func (nco NumConcurrentOption) SimOpt() {}

// Configures the simulation to ignore panics that occur during the execution of events.

// The simulation will not recover from panics that occur during the execution of events.
// This can be useful when debugging sections of the algorithm that panics.
// Default value is false
type IgnorePanicOption struct{}

func (ipo IgnorePanicOption) SimOpt() {}

// Configures the simulation to ignore errors that occur during the simulation of a run.

// Errors that occur during the simulation of a run will be ignored and the simulation of more runs will be continued.
// A summary of all the errors that occurred will be provided at the end.
// Useful to enure that the simulation continues even if some errors are found, allowing checking of the successful runs afterwards.
// Default value is false
type IgnoreErrorOption struct{}

func (ieo IgnoreErrorOption) SimOpt() {}
