package main

import (
	"experimentation/tester"
	"fmt"
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
	sch := tester.NewBasicScheduler[Rrb]()
	sm := tester.NewStateManager(
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
	tst := tester.NewSimulator[Rrb, State](sch, sm)
	tst.Simulate(
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
					tst.Send,
				)
			}
			return nodes
		},
		func(nodes map[int]*Rrb) error {
			nodes[0].Broadcast("Test Message")
			return nil
		},
	)
	fmt.Println(sch.EventRoot)
	fmt.Println(sm.StateRoot)

	fmt.Println(sm.StateRoot.Newick())
	fmt.Println(sch.EventRoot.Newick())

	checker := tester.NewPredicateChecker(
		tester.PredEventually(
			func(states map[int]State, terminal bool, _ []map[int]State) bool {
				// RB1: Validity
				for _, node := range states {
					for sentMsg := range node.sent {
						if !node.delivered[sentMsg] {
							return false
						}
					}
				}
				return true
			}),
		func(states map[int]State, terminal bool, _ []map[int]State) bool {
			// RB2: No duplication
			for _, node := range states {
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
		func(states map[int]State, terminal bool, _ []map[int]State) bool {
			// RB3: No creation
			sentMessages := map[message]bool{}
			for _, node := range states {
				for sent := range node.sent {
					sentMessages[sent] = true
				}
			}
			for _, state := range states {
				for delivered := range state.delivered {
					if !sentMessages[delivered] {
						return false
					}
				}
			}
			return true
		},
		tester.PredEventually(
			func(states map[int]State, terminal bool, _ []map[int]State) bool {
				// RB4 Agreement

				// Use leaf nodes to check for liveness properties
				// Can not say that the predicate has been broken for non-leaf nodes
				delivered := map[message]bool{}
				for _, node := range states {
					for msg := range node.delivered {
						delivered[msg] = true
					}
				}
				for msg := range delivered {
					for _, node := range states {
						if !node.delivered[msg] {
							return false
						}
					}
				}
				return true
			}),
	)

	resp := checker.Check(&sm.StateRoot)
	_, desc := resp.Response()
	print(desc)
}
