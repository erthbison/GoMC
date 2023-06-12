# Go-MC

A modular implementation level model checker for Go.

## Prerequisite

- Go: Version 1.19.4 or higher

## Learn More

- [How To Use](/Documentation/user-guide.md)
- [Configuration Options](/Documentation/configuration-guide.md)
- [Event Managers](/Documentation/event-managers.md)

## A (not so) Quick Showcase of Go-MC

### Simulating an Algorithm
To provide an example on how to use Go-MC to verify a algorithm we us the `Regular Reliable Broadcast` algorithm (Algorithm 3.3) from  `Introduction to reliable and secure distributed programming` by Cachin et al(2011). 
This examples omit some elements for brevity, the complete implementation and testing of the algorithm can be found under `examples/rrb`.

First, we decide what state should be collected from the nodes.
We do not collect all state, but focuses on collecting variables that will be used when verifying the algorithm. 
For example, we want to ensure that the algorithm does not create messages that was not sent (RB3: No Creation).
To do this we store a set of both delivered and sent messages and make sure that all messages that are in the delivered set is also in the sent set. 
We define a `struct` that stores the variables:

```go
type State struct {
	delivered      map[message]bool
	sent           map[message]bool
}
```

We then prepare the simulation. 
This includes configuring static options that will persist over multiple simulations, such as the State Manager and scheduler that will be used. 
The signature of the function used to prepare the simulation can be seen below:

```go 
func PrepareSimulation[T, S any](smOpts StateManagerOption[T, S], opts ...SimulatorOption) Simulation[T, S] 
```

The `StateManagerOption` is obligatory and is used to configure the State Manager that will be used. 
We configure the State Manager with a function that collects the specified state from the node and with a function that compares the equality of two states. 

```go
sim = gomc.PrepareSimulation(
    gomc.WithTreeStateManager(
        func(node *Rrb) State {
            return State{
                delivered:      maps.Clone(node.delivered),
                sent:           maps.Clone(node.sent),
                deliveredSlice: slices.Clone(node.deliveredSlice),
            }
        },
        func(s1, s2 State) bool {
            if !maps.Equal(s1.delivered, s2.delivered) {
                return false
            }
            if !slices.Equal(s1.deliveredSlice, s2.deliveredSlice) {
                return false
            }
            return maps.Equal(s1.sent, s2.sent)
        },
    ),
)
```

We can now run the simulation. 
When running the simulation we configure specific scenarios by defining the number of nodes, the request that will be sent to nodes and the checker that will be used and which nodes will fail during the simulation. 

```go
func (sr Simulation[T, S]) Run(InitNodes InitNodeOption[T], requestOpts RequestOption, checker CheckerOption[S], opts ...RunOptions) checking.CheckerResponse
```

The `InitNodeOption` is used to configure a function that will initialize the nodes for the simulation.
The function returns a `map[int]*T` where `T` is a node in the algorithm. 
The function is used to initialize the Event Manager that is inserted into the algorithm and allows the simulation to control the sources of randomness.

`RequestOption` configures a list of requests that will be sent to the nodes during the simulation. 
The requests are used to start the simulation.

`CheckerOption` defines the Checker that will be used to verify the properties of the algorithm. 
We use the `PredicateChecker` which uses a set of predicates, defines as go functions, to define the properties.

```go 
resp := sim.Run(
    gomc.InitNodeFunc(
        func(sp eventManager.SimulationParameters) map[int]*Rrb {
            // Initialize the event manager
            send := eventManager.NewSender(sp)
            nodes := map[int]*Rrb{}
            for _, id := range test.nodes {
                nodes[id] = NewRrb(
                    id,
                    test.nodes,
                    // insert the event manager into the node
                    send.SendFunc(id),
                )
            }
            return nodes
        },
    ),
    gomc.WithRequests(
        gomc.NewRequest(0, "Broadcast", "Test Message"),
    ),
    gomc.WithPredicateChecker(predicates...),
    gomc.WithPerfectFailureManager(
        func(t *Rrb) { t.crashed = true }, 
        test.crashedNodes...
    ),
)
```

The `WithPerfectFailureManager` option is an optional parameter that can be used to configure which nodes will crash during the simulation. 
The option is also configured with a function that performs the crash on the node.

```go
predicates = []checking.Predicate[State]{
    ...
    func(s checking.State[State]) bool {
		// RB3: No creation
		sentMessages := map[message]bool{}
		for _, node := range s.LocalStates {
			for sent := range node.sent {
				sentMessages[sent] = true
			}
		}
		for _, state := range s.LocalStates {
			for delivered := range state.delivered {
				if !sentMessages[delivered] {
					return false
				}
			}
		}
		return true
	},
    ...
}
```

We define the predicates that will be used by the `PredicateChecker`. 
The predicates are used to test that the desired properties holds.
They should return `true` when the property holds and `false` otherwise. 

### Running an Algorithm

Go-MC also offers support for running an algorithm live and collecting events that are executed. 
The `Runner` executes events on different nodes concurrently.
It is configured in much the same way as the simulation.

The `Runner` is prepared using the `PrepareRunner` function.
There are two mandatory options: `InitNodeOption`, which is similar to the one used to run a simulation and creates the nodes that will be used and `GetStateOption`, which is a function defining how to collect the state from the node. 
The `GetStateOption` is similar to one of the functions used to configure the State Manager. 

```go
func PrepareRunner[T, S any](initNodes InitNodeOption[T], getState GetStateOption[T, S], opts ...RunnerOption) *runner.Runner[T, S]
```

In the example below we configure an implementation of the paxos algorithm for running.
The algorithm can be found under `examples/paxos`.
Note that the complete initialization of the nodes is done in a separate function called `createNodes`. 

```go
r := gomc.PrepareRunner(
		gomc.InitNodeFunc(
			createNodes(addrMap)
		),
		gomc.WithStateFunction(func(t *paxos.Server) State {
			t.Lock()
			defer t.Unlock()
			return State{
				proposed: t.Proposal,
				decided:  t.Decided,
			}
		}),
		gomc.WithStopFunctionRunner(
            func(t *paxos.Server) { t.Stop() },
        ),
	)
```

The `PrepareRunner` function returns a configured and started `Runner`.
The `Runner` can then be given commands, such as sending requests to nodes, pausing and resuming the execution of events on a node, and crashing a node.

```go
err = r.Request(
    gomc.NewRequest(id, "Propose", "Value 1"),
)
```

