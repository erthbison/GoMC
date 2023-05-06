package main

import (
	"fmt"

	"testing"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"gomc"
	"gomc/checking"
	"gomc/eventManager"
)

var predicates = []checking.Predicate[State]{
	checking.Eventually(
		func(s checking.State[State]) bool {
			// RB1: Validity
			return checking.ForAllNodes(func(a State) bool {
				for sentMsg := range a.sent {
					if !a.delivered[sentMsg] {
						return false
					}
				}
				return true
			}, s, true)
		}),
	func(s checking.State[State]) bool {
		// RB2: No duplication
		for _, node := range s.LocalStates {
			delivered := make(map[message]bool)
			for _, msg := range node.deliveredSlice {
				if delivered[msg] {
					return false
				}
				delivered[msg] = true
			}
		}
		return true
	},
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
	checking.Eventually(
		func(s checking.State[State]) bool {
			// RB4 Agreement

			// Use leaf nodes to check for liveness properties
			// Can not say that the predicate has been broken for non-leaf nodes
			delivered := map[message]bool{}
			for _, node := range s.LocalStates {
				for msg := range node.delivered {
					delivered[msg] = true
				}
			}

			for msg := range delivered {
				if checking.ForAllNodes(func(s State) bool { return !s.delivered[msg] }, s, true) {
					return false
				}
			}
			return true
		},
	),
}

type State struct {
	delivered      map[message]bool
	sent           map[message]bool
	deliveredSlice []message
}

var sim gomc.Simulation[Rrb, State]

func init() {
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
		gomc.PrefixScheduler(),
	)
}

var simulations = []struct {
	nodes        []int
	crashedNodes []int
}{
	{
		[]int{0, 1, 2},
		[]int{},
	},
	{
		[]int{0, 1, 2},
		[]int{1},
	},
	{
		[]int{0, 1, 2, 4, 5, 6, 7, 8, 9},
		[]int{8},
	},
}

func TestRrb(t *testing.T) {
	for i, test := range simulations {
		resp := sim.Run(gomc.InitNodeFunc(
			func(sp gomc.SimulationParameters) map[int]*Rrb {
				send := eventManager.NewSender(sp.EventAdder)
				nodes := map[int]*Rrb{}
				for _, id := range test.nodes {
					nodes[id] = NewRrb(
						id,
						test.nodes,
						send.SendFunc(id),
					)
				}
				return nodes
			}),
			gomc.WithRequests(gomc.NewRequest(0, "Broadcast", "Test Message")),
			gomc.WithPredicateChecker(predicates...),
			gomc.WithPerfectFailureManager(func(t *Rrb) { t.crashed = true }, test.crashedNodes...),
		)
		_, desc := resp.Response()
		fmt.Printf("Test %v: Got result: %v\n", i, desc)
	}
}
