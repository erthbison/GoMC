package main

import (
	"fmt"
	"gomc"
	"gomc/scheduler"
	"strings"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type State struct {
	delivered      map[message]bool
	sent           map[message]bool
	deliveredSlice []message
}

func (s State) String() string {
	bldr := strings.Builder{}
	bldr.WriteString("Delivered: {")
	for _, key := range s.deliveredSlice {
		bldr.WriteString(fmt.Sprintf(" %v ", key))
	}
	bldr.WriteString("} Sent: {")
	for key := range s.sent {
		bldr.WriteString(fmt.Sprintf(" %v ", key))
	}
	bldr.WriteString("}")

	return bldr.String()
}

func main() {
	numNodes := 2
	sch := scheduler.NewBasicScheduler[Rrb]()
	sm := gomc.NewStateManager(
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
	)
	sender := gomc.NewSender[Rrb](sch)
	sim := gomc.NewSimulator[Rrb, State](sch, sm)
	err := sim.Simulate(
		func() map[int]*Rrb {
			nodeIds := []int{}
			for i := 0; i < numNodes; i++ {
				nodeIds = append(nodeIds, i)
			}
			nodes := map[int]*Rrb{}
			for _, id := range nodeIds {
				nodes[id] = NewRrb(
					id,
					nodeIds,
					sender.SendFunc(id),
				)
			}
			return nodes
		},
		map[int][]func(*Rrb) error{
			0: {
				func(node *Rrb) error {
					node.Broadcast("Test Message")
					return nil
				},
			},
		},
		[]int{},
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(sch.EventRoot)
	fmt.Println(sm.StateRoot)

	fmt.Println(sm.StateRoot.Newick())
	fmt.Println(sch.EventRoot.Newick())

	checker := gomc.NewPredicateChecker(
		gomc.PredEventually(
			func(states gomc.GlobalState[State], terminal bool, _ []gomc.GlobalState[State]) bool {
				// RB1: Validity
				for _, node := range states.LocalStates {
					for sentMsg := range node.sent {
						if !node.delivered[sentMsg] {
							return false
						}
					}
				}
				return true
			}),
		func(states gomc.GlobalState[State], terminal bool, _ []gomc.GlobalState[State]) bool {
			// RB2: No duplication
			for _, node := range states.LocalStates {
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
		func(states gomc.GlobalState[State], terminal bool, _ []gomc.GlobalState[State]) bool {
			// RB3: No creation
			sentMessages := map[message]bool{}
			for _, node := range states.LocalStates {
				for sent := range node.sent {
					sentMessages[sent] = true
				}
			}
			for _, state := range states.LocalStates {
				for delivered := range state.delivered {
					if !sentMessages[delivered] {
						return false
					}
				}
			}
			return true
		},
		gomc.PredEventually(
			func(states gomc.GlobalState[State], terminal bool, _ []gomc.GlobalState[State]) bool {
				// RB4 Agreement

				// Use leaf nodes to check for liveness properties
				// Can not say that the predicate has been broken for non-leaf nodes
				delivered := map[message]bool{}
				for _, node := range states.LocalStates {
					for msg := range node.delivered {
						delivered[msg] = true
					}
				}
				for msg := range delivered {
					for _, node := range states.LocalStates {
						if !node.delivered[msg] {
							return false
						}
					}
				}
				return true
			}),
	)

	resp := checker.Check(sm.StateRoot)
	_, desc := resp.Response()
	print(desc)
}
