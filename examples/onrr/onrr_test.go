package main

import (
	"gomc"
	"gomc/checking"
	"gomc/eventManager"
	"math"
	"testing"

	"golang.org/x/exp/slices"
)

type State struct {
	ongoingRead  bool
	ongoingWrite bool

	possibleReads []int

	read        int
	currentRead bool
}

func TestOnrr(t *testing.T) {
	// Select a scheduler. We will use the basic scheduler since it is the only one that is currently implemented
	// sch := gomc.NewBasicScheduler()
	sim := gomc.PrepareSimulation[onrr, State](
		gomc.RandomWalkScheduler(10000),
	)
	resp := sim.Run(
		gomc.InitNodeFunc(
			func(sp gomc.SimulationParameters) map[int]*onrr {
				numNodes := 5
				send := eventManager.NewSender(sp.EventAdder)

				nodeIds := []int{}
				for i := 0; i < numNodes; i++ {
					nodeIds = append(nodeIds, i)
				}
				nodes := make(map[int]*onrr)
				for _, id := range nodeIds {
					nodes[id] = NewOnrr(id, send.SendFunc(id), nodeIds)
				}
				go func() {
					for {
						<-nodes[0].WriteIndicator
					}
				}()
				return nodes
			},
		),
		gomc.WithRequests(
			gomc.NewRequest(0, "Write", 2),
			gomc.NewRequest(1, "Read"),
			gomc.NewRequest(2, "Read"),
			gomc.NewRequest(3, "Read"),
			gomc.NewRequest(4, "Read"),
		),
		gomc.WithTreeStateManager(
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
		),
		gomc.WithPredicate(
			checking.Eventually(
				func(s checking.State[State]) bool {
					// Check that all correct nodes have no ongoing reads or writes
					return checking.ForAllNodes(func(a State) bool { return !(a.ongoingRead || a.ongoingWrite) }, s, true)
				},
			),
			func(s checking.State[State]) bool {
				writer := 0
				possibleReadSlice := make([][]int, len(s.Sequence))
				for i, elem := range s.Sequence {
					possibleReadSlice[i] = elem.LocalStates[writer].possibleReads
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
				for id := range s.LocalStates {
					readStart := math.MaxInt
					for i, elem := range s.Sequence {

						node := elem.LocalStates[id]
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
							readStart = math.MaxInt
						}
					}
				}
				return true
			},
		),
		gomc.MaxRuns(10000),
	)
	ok, desc := resp.Response()
	if !ok {
		t.Errorf("Expected not to find an error in the implementation")
	}
	print(desc)
}
