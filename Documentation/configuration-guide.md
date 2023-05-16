# Configuration Options

Contents: 
- [Prepare Simulation Options](#prepare-simulation-options)
- [Run Options](#run-options)
- [Runner Options](#runner-options)

## Prepare Simulation Options

### StateManagerOption

Configure the StateManager used to manage the state of the distributed system.

The State Manager collects and manages the state of the system under testing.

#### `WithTreeStateManager[T, S any](getLocalState func(*T) S, statesEqual func(S, S) bool) StateManagerOption[T, S]`

Use a TreeStateManager in the simulation. 

The TreeStateManager organizes the state in a tree structure, which is stored in memory. The TreeStateManager is configured with a function collecting the local state from a node and a function checking the equality of two local states. 
#### ` WithStateManager[T, S any](sm stateManager.StateManager[T, S]) StateManagerOption[T, S]`
 
Use the provided state manger in the simulation.

### SchedulerOption
Configures the scheduler used by the simulation

The scheduler determines the method which will be used to explore the state space.
It determines the order that events will be executed in.
The default value is a `Prefix` Scheduler

#### `RandomWalkScheduler(seed int64) SimulatorOption`

Use a random walk scheduler for the simulation.

The random walk scheduler is a randomized scheduler.
It uniformly picks the next event to be scheduled from the currently enabled events.
It does not have a designated stop point, and will continue to schedule events until maxRuns is reached.
It does not guarantee that all runs have been tested, nor does it guarantee that the same run will not be simulated multiple times.
Generally, it provides a more even/varied exploration of the state space than systematic exploration

#### `PrefixScheduler() SimulatorOption`
Use a prefix scheduler for the simulation.

The prefix scheduler is a systematic tester, that performs a depth first search of the state space.
It will stop when the entire state space is explored and will not schedule identical runs.

#### `ReplayScheduler(run []event.EventId) SimulatorOption`

Use a replay scheduler for the simulation

The replay scheduler replays the provided run, returning an error if it is unable to reproduce it
The provided run is represented as a slice of event ids, and can be exported using the CheckerResponse.Export()

#### `WithScheduler(sch scheduler.GlobalScheduler) SimulatorOption`
Use the provided scheduler for the simulation

Used to configure the simulation to use a different implementation of scheduler than is commonly provided

### MaxDepthOption
Configures the max depth of a run

The depth of a run is the number of events that are executed in a run.
If a simulation reaches the maxDepth it will stop.
If a simulation is stopped early, we can not determine whether the algorithm is correct or not,
we can only determine whether we found some violations in the recorded states.
Decreasing the max depth will reduce the size of the state space.
This will reduce memory and computation time requirements.
Default value is `100`

#### `MaxDepth(maxDepth int) SimulatorOption`

Configure the maximum depth explored.

Default value is 100.

Note that liveness properties can not be verified if a run is not fully explored to its end.

### MaxRunsOption

Configures the maximum number of runs that will be simulated.

If the number of simulated runs reaches the maximum number of runs the simulation will stop.
If a simulation is stopped early, we can not determine whether the algorithm is correct or not,
we can only determine whether we found some violations in the recorded runs.
Decreasing the maximum number of runs will reduce the size of the discovered state space.
This will reduce memory and computation time requirements.
Default value is `10000`

#### `MaxRuns(maxRuns int) SimulatorOption`

Configure the maximum number of runs simulated

Default value is 10000

### NumConcurrentOption

Configures the number of runs that are simulated concurrently.

Higher values will reduce simulation time, but require more memory.
Default value is `GOMAXPROCS`

#### `NumConcurrent(n int) SimulatorOption`

Configure the number of runs that will be executed concurrently.

Default value is GOMAXPROCS

### IgnorePanicOption

Configures the simulation to ignore panics that occur during the execution of events.

The simulation will not recover from panics that occur during the execution of events.
This can be useful when debugging sections of the algorithm that panics.
Default value is `false`

#### `IgnorePanic() SimulatorOption`

Set the ignorePanic flag to true.

If true will ignore panics that occur during the simulation and let them execute as normal, stopping the simulation.
If false will catch the panic and return it as an error.
Ignoring the panic will make it easier to troubleshoot the error since you can use the debugger to inspect the state when it panics. It will also make the simulation stop.

### IgnoreErrorOption

Configures the simulation to ignore errors that occur during the simulation of a run.

Errors that occur during the simulation of a run will be ignored and the simulation of more runs will be continued.
A summary of all the errors that occurred will be provided at the end.
Useful to enure that the simulation continues even if some errors are found, allowing checking of the successful runs afterwards.
Default value is `false`

#### `IgnoreError() SimulatorOption`

Set the ignoreError flag to true.

If true will ignore all errors while simulating runs. Will return aggregate of errors at the end.
If false will interrupt simulation if an error occur.

## Run Options

### InitNodeOption

Configures how the nodes are started.

The function should create the nodes that will be used when running the simulation.
It should also initialize the Event Manager that will be used.
The provided SimulationParameters should be used to configure the Event Managers with run specific data. 

#### `InitNodeFunc[T any](f func(sp eventManager.SimulationParameters) map[int]*T) InitNodeOption[T]`

Uses the provided function f to generate a map of the nodes.

#### `InitSingleNode[T any](nodeIds []int, f func(id int, sp eventManager.SimulationParameters) *T) InitNodeOption[T]`

Uses the provided function f to generate individual nodes with the provided id and add them to a map of the nodes.

### RequestOption

Configures the requests to the distributed system.

The request are used to start the simulation and define the scenario of the simulation.

#### `WithRequests(requests ...request.Request) RequestOption`

Configures the requests to the distributed system.

The request are used to start the simulation and define the scenario of the simulation.

### CheckerOption

Configures the Checker to be used when verifying the algorithm.

The Checker verifies that the properties of the algorithm holds.
It returns a CheckerResponse with the result of the simulation.

#### `WithPredicateChecker[S any](predicates ...checking.Predicate[S]) CheckerOption[S]`

Use a PredicateChecker to verify the algorithm.

The predicate checker uses functions to define the properties of the algorithm.
The functions are provided as the checking.Predicate type.

#### `WithChecker[S any](checker checking.Checker[S]) CheckerOption[S]`

Specify the Checker used to verify the algorithm.

### FailureManagerOption

Configures the Failure Manager that will be used during the simulation

The Failure Manager controls what crash abstractions will be represented by the simulation 
and which nodes will crash at some point during the simulation.
The Failure Manager also works as a Failure Detector that informs nodes about status changes.
Different Failure Managers can represent different types of Failure Detectors.
Default value is no node crashes.

#### `WithFailureManager[T any](fm failureManager.FailureManger[T]) RunOptions`

Specify the failure manager used for the Simulation

Default value is a PerfectFailureManager with no node crashes.

#### `WithPerfectFailureManager[T any](crashFunc func(*T), failingNodes ...int) RunOptions`

Configure the simulation to use a PerfectFailureManager.

The PerfectFailureManager implements the crash-stop failures in a synchronous system.
It imitates the behavior of the Perfect Failure Detector.

Default value is a PerfectFailureManager with no node crashes.

### ExportOption

Configures io.writers that the discovered state will be exported to

Can be applied multiple times to add multiple io.writers.
Default value is no writers. 

#### `Export(w io.Writer) RunOptions`

Add a writer that the state will be exported to

Can be called multiple times.
Default value is no writers

### StopOption

Configures a function to shut down a node after the execution of a run.

The function should clean up any operations to avoid memory leaks across runs.
Default value is an empty function.

#### `WithStopFunctionSimulator[T any](stop func(*T)) RunOptions`

Configures a function used to stop the nodes after a run.

The function should clean up all operations of the nodes to avoid memory leaks across runs.

Default value is empty function.

## Runner Options

### InitNodeOption

Configures how the nodes are started.

The function should create the nodes that will be used when running the simulation.
It should also initialize the Event Manager that will be used.
The provided SimulationParameters should be used to configure the Event Managers with run specific data. 

#### `InitNodeFunc[T any](f func(sp eventManager.SimulationParameters) map[int]*T) InitNodeOption[T]`

Uses the provided function f to generate a map of the nodes.

#### `InitSingleNode[T any](nodeIds []int, f func(id int, sp eventManager.SimulationParameters) *T) InitNodeOption[T]`

Uses the provided function f to generate individual nodes with the provided id and add them to a map of the nodes.

### GetStateOption

Configure how to collect the local state from a node

#### `WithStateFunction[T, S any](f func(*T) S) GetStateOption[T, S]`

Configure how to collect the local state from a node

#### `WithStopFunctionRunner[T any](stop func(*T)) RunnerOption`

Configures a function used to stop the nodes after a run.

The function should clean up all operations of the nodes to avoid memory leaks across runs.

Default value is empty function.

### RecordChanBufferOption

Configures how many records can be buffered

Default value is `100`

#### `RecordChanSize(size int) RunnerOption`

Configure the buffer size of the records channel

Default value is 100

### EventChanBufferOption

Configures how many pending event can be buffered by each node

Default value is `100`

#### `EventChanBufferSize(size int) RunnerOption`

Configure the buffer size of the event channel used to store pending events

Default value is 100