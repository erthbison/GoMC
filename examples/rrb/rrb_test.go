package main

import (
	"fmt"

	"strings"
	"testing"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"gomc"
	"gomc/checking"
	"gomc/eventManager"
)

type State struct {
	delivered      map[message]bool
	sent           map[message]bool
	deliveredSlice []message
}

func (s State) String() string {
	bldr := &strings.Builder{}
	bldr.WriteString("D:{")
	for _, key := range s.deliveredSlice {
		fmt.Fprintf(bldr, "%v,", key)
	}
	bldr.WriteString("}S:{")
	for key := range s.sent {
		fmt.Fprintf(bldr, "%v,", key)
	}
	bldr.WriteString("}")
	return bldr.String()
}

func TestRrb(t *testing.T) {
	numNodes := 5

	sim := gomc.PrepareSimulation[Rrb, State](
		gomc.PrefixScheduler(),
	)

	resp := sim.Run(gomc.InitNodeFunc(
		func(sp gomc.SimulationParameters) map[int]*Rrb {
			send := eventManager.NewSender(sp.EventAdder)
			nodeIds := []int{}
			for i := 0; i < numNodes; i++ {
				nodeIds = append(nodeIds, i)
			}
			nodes := map[int]*Rrb{}
			for _, id := range nodeIds {
				nodes[id] = NewRrb(
					id,
					nodeIds,
					send.SendFunc(id),
				)
			}
			return nodes
		}),
		gomc.WithRequests(gomc.NewRequest(0, "Broadcast", "Test Message")),
		gomc.WithTreeStateManager(
			func(node *Rrb) State {
				newDelivered := map[message]bool{}
				for key, value := range node.delivered {
					newDelivered[key] = value
				}
				newSent := map[message]bool{}
				for key, value := range node.sent {
					newSent[key] = value
				}
				newDeliveredSlice := make([]message, len(node.deliveredSlice))
				copy(newDeliveredSlice, node.deliveredSlice)
				return State{
					delivered:      newDelivered,
					sent:           newSent,
					deliveredSlice: newDeliveredSlice,
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
		gomc.WithPredicateChecker(
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
						for _, node := range s.LocalStates {
							if !node.delivered[msg] {
								return false
							}
						}
					}
					return true
				},
			),
		),
	)
	ok, desc := resp.Response()
	if ok {
		t.Errorf("Expected to find an error in the implementation")
	}
	print(desc)
}
