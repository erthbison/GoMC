package main

import (
	"gomc"
	"gomc/scheduler"

	"golang.org/x/exp/slices"
)

const MaxInt = int(^uint(0) >> 1)

type State struct {
	ongoingRead  bool
	ongoingWrite bool

	possibleReads []int

	read        int
	currentRead bool
}

func main() {
	// Select a scheduler. We will use the basic scheduler since it is the only one that is currently implemented
	// sch := gomc.NewBasicScheduler[onrr]()
	sch := scheduler.NewRandomScheduler[onrr](10000)

	// Configure the state manager. It takes a function returning the local state of a node and a function that checks for equality between two states
	sm := gomc.NewStateManager(
		func(node *onrr) State {
			reads := []int{}
			reads = append(reads, node.possibleReads...)

			// If there has been a read indication store it. Otherwise ignore it
			read := 0
			currentRead := false
			select {
			case read = <-node.ReadIndicator:
				currentRead = true
			default:
			}

			return State{
				ongoingRead:   node.ongoingRead,
				ongoingWrite:  node.ongoingWrite,
				possibleReads: reads,
				read:          read,
				currentRead:   currentRead,
			}
		},
		func(a, b State) bool {
			if a.ongoingRead != b.ongoingRead {
				return false
			}
			if a.ongoingWrite != b.ongoingWrite {
				return false
			}
			if a.currentRead != b.currentRead {
				return false
			}
			if a.read != b.read {
				return false
			}
			return slices.Equal(a.possibleReads, b.possibleReads)
		},
	)

	// Create a simulator. providing a function specifying how to instantiate the nodes and a function specifying how to start the test
	simulator := gomc.NewSimulator[onrr, State](sch, sm)
	sender := gomc.NewSender[onrr](sch)
	err := simulator.Simulate(
		func() map[int]*onrr {
			numNodes := 5

			nodeIds := []int{}
			for i := 0; i < numNodes; i++ {
				nodeIds = append(nodeIds, i)
			}
			nodes := make(map[int]*onrr)
			for _, id := range nodeIds {
				nodes[id] = NewOnrr(id, sender.SendFunc(id), nodeIds)
			}
			return nodes
		},
		map[int][]func(*onrr) error{
			0: {
				func(node *onrr) error {
					go func() {
						// ignore the write indication
						<-node.WriteIndicator
					}()
					node.Write(2)
					return nil
				},
			},
			1: {
				func(node *onrr) error {
					node.Read()
					return nil
				},
			},
			2: {
				func(node *onrr) error {
					node.Read()
					return nil
				},
			},
			3: {
				func(node *onrr) error {
					node.Read()
					return nil
				},
			},
			4: {
				func(node *onrr) error {
					node.Read()
					return nil
				},
			},
		},
		[]int{3, 4},
	)

	if err != nil {
		panic(err)
	}

	// fmt.Println(sch.EventRoot)
	// fmt.Println(sm.StateRoot)

	// fmt.Println(sm.StateRoot.Newick())
	// fmt.Println(sch.EventRoot.Newick())

	checker := gomc.NewPredicateChecker(
		gomc.PredEventually(
			func(states map[int]State, _ bool, _ []map[int]State) bool {
				for _, state := range states {
					if state.ongoingRead || state.ongoingWrite {
						return false
					}
				}
				return true
			},
		),
		func(state map[int]State, _ bool, seq []map[int]State) bool {
			writer := 0
			possibleReadSlice := make([][]int, len(seq))
			for i, elem := range seq {
				possibleReadSlice[i] = elem[writer].possibleReads
			}

			// Create a set of the set of possible values for an event at the provided range
			// Possible values include the union of all locally stored possible value for all states in the range
			possibleVals := func(start, end int) map[int]bool {
				possibleReads := map[int]bool{}
				for i := start; i <= end; i++ {
					for _, val := range possibleReadSlice[i] {
						possibleReads[val] = true
					}
				}
				return possibleReads
			}

			// For each node in. Go trough the sequence and find ReadEvents.
			// Find possible values for the read event and check that it matches the returned value
			for id := range state {
				readStart := MaxInt
				for i, elem := range seq {

					node := elem[id]
					if node.ongoingRead && i < readStart {
						// A read operation starts
						readStart = i
					}
					if node.currentRead {
						// A read operation ends
						valSet := possibleVals(readStart, i)
						if !valSet[node.read] {
							return false
						}
						readStart = MaxInt
					}
				}
			}
			return true
		},
	)

	resp := checker.Check(sm.StateRoot)
	_, desc := resp.Response()
	print(desc)
}
