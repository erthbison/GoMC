# Implementing the (1, N) Regular Register for Tester

We will implement the (1, N) Regular Register for the testing. The process is divided into two steps. First we will implement the register. Then we will set up the tester with the register to verify the implementation.

## (1, N) Regular Register

We will use the (1, N) Regular Register(also called onrr), module 4.1 in [1]. The module describes a register with 1 designated writer and N designated readers. Every read operation returns the value of the last value written. A read operation that is concurrent with a write operation may return the last value written or the concurrently written value. 

To avoid the use of a perfect failure detector we will use algorithm 4.2, Majority Voting Regular Register from [1]. 

We first define the `onrr` struct with the variables defined in the algorithm. We define a `WriteIndicator` and a `ReadIndicator` which are channels used to send indications that a write and read operation has completed. Since the program will be run sequentially during simulation the channels must be buffered. Note that when we say that the program will run sequentially we primarily refer to the fact that the tester runs a loop of stages that will be run sequentially. First it will execute execute an event, then it will store the global state. If the channels are not buffered the program will block while executing the event. The nodes can use multiple goroutines, but it should be deterministic in the sense that calling a function several times with the same state should return the same result. 

We also store some variables that will be used to verify the algorithm. These are not used in the algorithm. The `ongoingRead` and `ongoingWrite` variables indicate wether the node is currently processing a `Read` or `Write` operation. They will be updated during the execution of the algorithm to ensure that they are correct at all times. `possibleReads` is a slice of all the values that are valid results for the `Read` operations. This is the last written value and the concurrently written value if it exists. 

In addition we also define a `nodes` slice containing the ids of all the nodes in the system, a `id` field storing the id of the current node and a `send` function that is used to represent a point-to-point link and will take care of all the communication between nodes. The function has the following signature: 

```go
func(from, to int, msgType string, msg []byte)
```

The function is used by the tester to order message arrival and explore all possible states that can occur in the algorithm. We will return to how this function is used later, when we define the methods of the `onrr`. 

We define a `NewOnrr` to instantiate an instance of the `onrr` type according to the algorithm. It takes the slice of all node ids and the `send` function as arguments. 

We will also define several structs representing the messages used by the algorithm. These are `BroadcastWriteMsg`, `AckMsg`, `BroadcastReadMsg` and `ReadValueMsg`, and they contain the fields defined in the algorithm. They contain no methods and are purely used for structuring the data. We create helper functions `encodeMsg` and `decodeMsg` that will encode and decode between structs and `[]byte`.

We then continue by defining the `Write` function, which is used to invoke a write operation on the register. It increments the `wts` variable and resets the `acks` variable according to the algorithm. It then generates a message and sends it to every node using the `send` function. 

The `send` function takes the `id` of the sender and the receiver as first and second argument. As third argument it takes the message type which is a `string` representing the name of method that will be called by the receiver to handle this message. This method is currently not defined, but we will call it `BroadcastWrite`. This function must be exported to ensure that it can be called with the message on the receiver. The fourth argument is a `[]byte` for the message that is going to be sent. 

We then define the `BroadcastWrite` method according to the algorithm. Since this method is a receiver of a network message it must have the following signature:

```go
func(from int, to int, msg []byte)
```

The first argument is the `id` of the sender. The second argument is the `id` of the receiver. The third argument is the message as a `[]byte`. The method acknowledges the write broadcast by generating a `AckMsg` and sending it with the `send` function. 

The remaining methods are implemented according to the algorithm, using the `send` function for all message transfers and referring to the corresponding message handler function. 

## Configuring the tester

The tester is configured in the `main.go` file. First a `State` struct is defined. It will be used to define the local state stored by the tester and will contain the information required to ensure that the properties of the module holds for all possible states of the implementation. One state instance is stored for each node for each event in the tester. The local state of all the nodes is combined in a map, `map[int]State`, to form the global state after the set event. 

To decide how to define the `State` we need to look at the properties we want to verify. These are *Termination*, informally stating that each operation that is started is eventually completed, and *Validity*, specifying the value that will be read according to the *regular register* abstraction. 

To verify the *Termination* property we use the `ongoingRead` and `ongoingWrite` variables. These will be updated by the nodes as the algorithm run. To verify *Validity* we store the set of possible values for each state. The set is maintained by the writer node. We also store a `read` and a `currentRead` variable. `read` is the value returned by the `ReadIndicator` at the current event, if there was such a value. `currentRead` indicates whether there was a value returned by the `ReadIndicator` at the current event. 

We then select a `Scheduler` to be used and define the `StateManager`. Currently the only scheduler available is the `BasicScheduler`. The `StateManager` requires two arguments; a function specifying how to create the local state, the `State` struct, from a node and a function specifying how to check states for equality. 

We then create a `Simulator` with the chosen `Scheduler` and `StateManager`. We call the `Simulate` method to start the simulation. It takes two arguments: a function specifying how to create and initialize the nodes and a function specifying the start action the nodes should perform. This should be some combinations of the actions performed by the system. 

We then define a `Checker` and provides functions that takes an instance of the global state and checks that the properties hold for this instance. The functions should return `true` if it holds and `false` id it breaks the property. The `Check` function iterates over all states that occurred during the simulation and verifies that all the functions returns `true` for the state. If it does not it returns information about which function failed and the sequence of states that resulted in the broken predicate. 


<!-- What functions do we use to verify the properties ??? -->
To create a function that verifies the *Termination* property we can simply check that the `ongoingRead` and `ongoingWrite` variables are false for all nodes. Since the property states that it should eventually complete, we should not check all global states in the execution. Therefore we only check terminal nodes. This can be done by checking the `terminal` variable or by using the `PredEventually` function.

To verify the *Validity* we check that any value returned by the `ReadIndicator`, i.e. the value stored in `read`, is in the set of possible reads. 

To run the checker we call the run function with the root of the state tree as parameter. 


<!-- 
Notes:
    - hard to define predicates. In particular it is hard to check the returned values in the moment they where returned. Found a solution where we check the channel in the getLocalState function. Requires the channel to be buffered (or for it to be filled in a goroutine i suppose, but that might be weird)

    - can not specify multiple start functions and ensure that all orders of them are tried. Requires that the user think more about how the start functions work. Could be solved by creating an event for function calls

    - Tedious to convert messages into `[]byte` before sending it. Could do it automatically using glob. Could be harder to do the decoding, but we might be able to use reflection to find the expected type and use that. 

    - Output from the checker is not that useful in identifying what happened. We should also create a way to export and import the states tree so that we dont have to run the simulation every time we want to run the checker. Can be done by implementing a NewickRead function
 -->