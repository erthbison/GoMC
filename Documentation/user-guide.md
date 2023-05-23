# User Guide

Go-MC is a tool for verifying and visualizing the implementation of distributed systems. 
Go-MC provides two functionalities: The Simulator, and The Runner. 
The simulator verifies the algorithm by controlling the execution of events, while the Runner executes the algorithm live and records events and state changes of the system.

## The Simulator

The simulator is used to verify that an implementation of a distributed algorithm operates as expected.
Simulating a distributed algorithm consists of two parts: 
1) Configuring the Simulation
2) Running the Simulation

Before starting the simulation of the algorithm, it should be noted that the simulation of a distributed system is a costly process and the state space explosion problem ensures that the memory and computation requirements quickly increases as the complexity of the distributed system increases. 
Some suggestions for reducing the resources required to run the simulation is provided at the end. 

This guide mentions some important options. For a full overview of all possible options see the [configuration guide](/Documentation/configuration-guide.md).

### Configuring the Simulation

The simulation is configured using the `PrepareSimulation` function.

```go
func PrepareSimulation[T, S any](smOpts StateManagerOption[T, S], opts ...SimulatorOption) Simulation[T, S]
```

The `T` generic type represent a node of the algorithm.
All nodes in the distributed system must run the same local algorithm, i.e. they must all be of the same type.

The `S` generic type represent the local state of a node. 
The type is used to configure which parts of the state of the node will be stored by Go-MC.
An example of a local state is shown below. 

```go 
type state struct {
	proposed Value[int]
	decided  []Value[int]
}
```

The `StateManagerOption` is used to configure the **State Manager** that will be used for the simulation.
The **State Manager** is configured with a function that collects the local state of a node and a function that checks the equality of two states. 

```go
gomc.WithTreeStateManager(
	func(node *HierarchicalConsensus[int]) state {
		return state{
			proposed: node.ProposedVal,
			decided:  slices.Clone(node.DecidedVal),
		}
	},
	func(a, b state) bool {
		if a.proposed != b.proposed {
			return false
		}
		return slices.Equal(a.decided, b.decided)
	},
)
```

In addition to configuring the **State Manager**, you can also specify the **Scheduler** that will be used when running the simulation. 
The scheduler is responsible for deciding the order in which events are executed in a run and for ensuring that the state space is properly explored.
The choice of **Scheduler** has a significant impact on the performance and results of the simulation. 
The **Scheduler** can also be used to replay previously executed runs.
The **Scheduler** can be configured by using a `SchedulerOption`, for a full list of the available options see the [configuration guide](/Documentation/configuration-guide.md#scheduleroption)

### Running the Simulation

The `PrepareSimulation` function returns an instance of the `Simulation` type.
This type can be used to run multiple simulations by calling the `Run` method. 
The different parameters are used to define different scenarios for the verification of the algorithm. 
Note that only one simulation can be run at the same time. 

```go
func (sr Simulation[T, S]) Run(InitNodes InitNodeOption[T], requestOpts RequestOption, checker CheckerOption[S], opts ...RunOptions) checking.CheckerResponse
```

The `Run` method has three mandatory parameters. 
The `InitNodeOption` is used to specify a function that will be used to initialize the nodes used in the simulation.
The function is also used to initialize the **Event Managers** with the provided `SimulationParameters` for the run.

```go
gomc.InitSingleNode(nodeIds,
	func(id int, sp eventManager.SimulationParameters) *HierarchicalConsensus[int] {
		send := eventManager.NewSender(sp)
		node := NewHierarchicalConsensus[int](
			id,
			nodeIds,
			send.SendFunc(id),
		)
		sp.CrashSubscribe(id, node.Crash)
		return node
	},
)
```

The `RequestOption` is used to configure a set of requests that will be made to the nodes during the simulation of the algorithm.
A request can be created using the `NewRequest` function and specifying the id of the target node, the name of the method that should be called on the node and the parameters that should be passed to the method. 
The requests will be added to the simulation and interleaved along with the other events, ensuring that all combinations of messages and requests are simulated. 
At least one valid request must be provided to the system. 
Different requests represent different scenarios, which can induce different types of errors in the distributed system.
It is therefore important to test with different combinations of requests.

```go
gomc.WithRequests(
	gomc.NewRequest(1, "Propose", Value[int]{1}),
	gomc.NewRequest(2, "Propose", Value[int]{2}),
	gomc.NewRequest(3, "Propose", Value[int]{3}),
)
```

```go
func NewRequest(id int, method string, params ...any) request.Request
```

The `CheckerOption` configures the **Checker** that will be used to verify the properties of the distributed system. 
The **Checker** is run after the simulations have been completed and it verifies that the properties holds for all states that was discovered.
The only available **Checker** is the `PredicateChecker`, where properties are defined as functions, of the type `Predicate`, returning `true` if the property hold and `false` if it is violated. 

```go
gomc.WithPredicateChecker(
	checking.Eventually(
		// C1: Termination
		func(s checking.State[state]) bool {
			return checking.ForAllNodes(func(s state) bool {
				return len(s.decided) > 0
			}, s, true)
		},
	),
	func(s checking.State[state]) bool {
		// C2: Validity
		proposed := make(map[Value[int]]bool)
		for _, node := range s.LocalStates {
			proposed[node.proposed] = true
		}
		return checking.ForAllNodes(func(s state) bool {
			if len(s.decided) < 1 {
				// The process has not decided a value yet
				return true
			}
			return proposed[s.decided[0]]
		}, s, false)
	},
	func(s checking.State[state]) bool {
		// C3: Integrity
		return checking.ForAllNodes(func(s state) bool { return len(s.decided) < 2 }, s, false)
	},
	func(s checking.State[state]) bool {
		// C4: Agreement
		decided := make(map[Value[int]]bool)
		checking.ForAllNodes(func(s state) bool {
			for _, val := range s.decided {
				decided[val] = true
			}
			return true
		}, s, true)
		return len(decided) <= 1
	},
)
```
The `Predicates` have one parameter: a variable of the `State` type. 
The `State` stores the state of a system at a certain point of the simulation.
It also stores the sequence of `GlobalState`s that represent the current run, including the current state. 

```go 
type Predicate[S any] func(s State[S]) bool
```

```go
// The state of the system at the current point of execution
type State[S any] struct {
	// The local states of the nodes.
	LocalStates map[int]S
	// The status of the nodes. True means that the node is correct, false that it has crashed.
	Correct map[int]bool
	// True if this is the last recorded state in a run. False otherwise.
	IsTerminal bool
	// The sequence of GlobalStates that lead to this State.
	Sequence []state.GlobalState[S]
}

```

The helper functions `Eventually` and `ForAllNodes` are also provided to simplify the process of defining predicates. 

The optional configuration `FailureManagerOption` can be used to configure a **Failure Manager** that will be used during the simulation.
The **Failure Manager** determines the failure abstraction that will be supported by the simulation.
It can also provide the functionality of a failure detector.
Go-MC currently only provides the `PerfectFailureManager` which supports the crash-stop failure abstraction with a perfect failure detector, but more **Failure Managers** can be implemented. 

By default the **Failure Manager** performs no node crashes, but you can configure it to crash specified nodes during the simulation.
The `PerfectFailureManager` is configured with a `crash` function specifying how the node crashes and a list of the nodes that will crash during the simulation. 
The `crash` function must ensure that the node does not respond to new messages or requests. 
The list of nodes that will crash during the simulation is provided as a list of the ids of the nodes.
The crashes of nodes will be interleaved with other events, such as messages arriving or requests, to ensure that all possible timings are verified. 

```go
gomc.WithPerfectFailureManager(
	func(t *HierarchicalConsensus[int]) { t.crashed = true },
	1,
),
```

If the nodes that are created starts separate goroutines, for example when listening to incoming messages, a stop function must be run after each run to clean up after the node.
This ensures that the nodes of the run is properly stopped and that no memory leaks occur when they are discarded.
The stop function is configured with a `StopOption`.
By default, no stop function is used.

```go
gomc.WithStopFunctionSimulator(
	func(t *GrpcConsensus) { t.Stop() },
)
```

### Adapting the Algorithm

<!-- TODO: Perhaps say something about designing the algorithm? that some changes must be made for the Event Managers, and that in general the simulation makes some assumptions? -->

While it is desirable for Go-MC to be usable when verifying any implementation of an algorithm out of the box, some basic assumptions must be made to be able to simulate the algorithm.

Firstly, Go-MC assumes that once an event is started it will continue executing until it is completed. 
An event can not pause its execution and wait for some other event to be processed.
If this assumption is broken the simulation will not progress beyond the event. 

Secondly, it is assumed that all events are deterministic. 
This is required for most **Schedulers** to be able to build a view of the state space.

Finally, there must be a away for the **Event Managers** to be inserted into the algorithm. 
The specifics of this varies between **Event Managers**, but it might require minimal changes to the implementation.
The changes should be small and care should be taken to ensure that they do not affect the outcome of the algorithm, to preserve the validity of the results.

### Reducing the Required Resources
The complexity of the algorithm has a significant impact on the resources required of the simulation.
Complex algorithms consists of more possible runs, and will therefore take longer time to simulate.
They also requires more memory to maintain the discovered state space. 
The complexity is generally measured by the number of events are simulated, also called the depth of the simulation.

In general, messages are one of the most common types of events.
One way of keeping the complexity of the algorithm low is therefore to reduce the number of messages sent in the algorithm.
This is already good practice, since messages are considered time-consuming and expensive.
Optimizing the number of messages that are sent, is therefore a good way of reducing the resource requirements.

One way of reducing the number of events that occur is to reduce the number of nodes in the simulation.
Large distributed systems will naturally produce more events than small systems, while not necessarily increasing the probability of finding a known error. 
It is therefore useful to design small scenarios with a limited number of nodes, to ensure that the simulation converges quickly and can cover a large part of the state space. 

We also recommend taking a modular approach when testing distributed systems. 
By dividing the algorithm into several modules that can be simulated independently the complexity of each simulation can remain low, even if the complexity of the system is high.
Custom **Event Managers** can be created to mock modules during the simulation.
For more details about the **Event Managers**, see [event-managers.md](/Documentation/event-managers.md).
Mocking modules in this way will have a large impact on the complexity of the simulation, since complex modules that normally would require multiple rounds of messages can be modelled by using a single event. 

In practice, it is not always possible to reduce the complexity of the simulation to a level where the entire state space can be explored.
In these scenarios we can limit the number of runs that are simulated or the maximum depth of the runs that are simulated.
This will allow us to explore parts of the state space.
This will naturally impact the simulations ability to find existing bugs, since there is no guarantee that the bug is in the part of the state space that is explored.
It will however provide some level of assurance that the most common bugs have been eliminated.

The [`MaxRunsOption`](/Documentation/configuration-guide.md#maxrunsoption) option is used to configure the maximum number of runs that will be simulated. 
Configuring a low value (around 1000) will significantly reduce the memory and computation requirements of the simulation.
It can however have a significant impact on the probability of finding an existing bug in the implementation, particularly when using certain **Schedulers**. 

The maximum depth of a simulation can be configured with the [`MaxDepthOption`](/Documentation/configuration-guide.md#maxdepthoption). 
Reducing the maximum depth will result it the simulation of runs stopping early. 
This reduces the size of the state space, but also limits the ability to verify the Liveness properties of the simulation.

The choice of **Scheduler** can have an significant impact on the resources required by and the performance of the simulation. 
It is therefore important to consider the properties of the specific **Scheduler** to ensure that the correct one is chosen for the simulation.
It is also important to consider the maximum depth and number of runs when selecting the **Scheduler**.
Some **Schedulers** operates by iteratively changing the run that is simulated. 
Such a **Schedulers** end up simulating similar runs, which can be a problem when the size of the explored state space is small, compared to the size of the actual state space.
In such cases it might be better to use a **Scheduler** that more evenly samples runs from the entire state space, to increase the chances of finding different bugs.

## The Runner

The Runner is used to record the execution of the algorithm during execution.
It collects records of events and states from the nodes in the system and sends them to the subscribed user.

```go
func PrepareRunner[T, S any](initNodes InitNodeOption[T], getState GetStateOption[T, S], opts ...RunnerOption) *runner.Runner[T, S]
```

The Runner is created by the `PrepareRunner` function.
The function prepares and starts a Runner with the provided configuration.
In the same way as the Simulator, the Runner  uses two generic variables: `T` and `S`.
`T` specifies the type implementing the node and `S` specifies the type storing the local state of a node.
 
The `InitNodeOption` is used to specify a function that will be used to initialize the nodes.
This is similar to the option used when running the simulation. 

```go
gomc.InitSingleNode(nodeIds,
	func(id int, sp eventManager.SimulationParameters) *HierarchicalConsensus[int] {
		send := eventManager.NewSender(sp)
		node := NewHierarchicalConsensus[int](
			id,
			nodeIds,
			send.SendFunc(id),
		)
		sp.CrashSubscribe(id, node.Crash)
		return node
	},
)
```

The `GetStateOption` configures a function used to collect the state from the nodes. 
This is similar to the function used to configure the **State Manager** in the simulation. 

```go
gomc.WithStateFunction(
	func(t *paxos.Server) State {
		return State{
			proposed: t.Proposal,
			decided:  t.Decided,
		}
	},
)
```

The `StopOption` is optional, but must be provided if you want to trigger a crash of a node during the running.
The option configures a function that will be used to perform the crash of a node, similar to the function provided to the **Failure Manager** when configuring the run of the Simulation.

```go
gomc.WithStopFunctionRunner(
	func(t *paxos.Server) { t.Stop() },
)
```

You can interact with the distributed system through the provided Runner.
The following commands are provided by the Runner: 
- `Request`: Add a request to the system
- `PauseNode`: Pause the execution of events on the specified node
- `ResumeNode`: Resume the execution of events on the specified node
- `CrashNode`: Trigger a crash of the specified node

You can subscribe to receive records collected by the Runner through the `SubscribeRecords` method.
The runner sends three types of records:
- `ExecutionRecord` are sent when an internal event is executed.
- `MessageRecord` are sent when a message is sent or received by a node.
They contain information about the message. 
- `StateRecord` are sent after an event has been executed and contains the new state of the node. 
