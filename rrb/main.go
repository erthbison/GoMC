package main

import (
	"experimentation/tester"
	"fmt"
	"strings"

	"golang.org/x/exp/maps"
)

type State struct {
	delivered map[message]bool
	sent      map[message]bool
}

func (s State) String() string {
	bldr := strings.Builder{}
	bldr.WriteString("Delivered: {")
	for key := range s.delivered {
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
	sch := tester.NewBasicScheduler()
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
			return State{
				delivered: newDelivered,
				sent:      newSent,
			}
		},
		func(s1, s2 State) bool {
			if !maps.Equal(s1.delivered, s2.delivered) {
				return false
			}
			return maps.Equal(s1.sent, s2.sent)
		},
	)
	tst := tester.CreateTester[Rrb, State](sch, sm)
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
		func(nodes map[int]*Rrb) {
			for i := 0; i < 1; i++ {
				nodes[0].Broadcast(fmt.Sprintf("Test Message - %v", i))
			}
		},
	)
	fmt.Println(sch.EventRoot)
	fmt.Println(sm.StateRoot)

	fmt.Println(sm.StateRoot.Newick())
	fmt.Println(sch.EventRoot.Newick())

	checker := tester.NewPredicateChecker(
		func(states map[int]State, leaf bool) (bool, string) {
			// RB3: No creation
			desc := "RB3: No Creation"
			sentMessages := map[message]bool{}
			for _, node := range states {
				for sent := range node.sent {
					sentMessages[sent] = true
				}
			}
			for _, state := range states {
				for delivered := range state.delivered {
					if !sentMessages[delivered] {
						return false, desc
					}
				}
			}
			return true, desc
		},
		func(states map[int]State, leaf bool) (bool, string) {
			// RB1: Validity
			desc := "RB1: Validity"
			if !leaf {
				return true, desc
			}
			for _, node := range states {
				for sentMsg := range node.sent {
					if !node.delivered[sentMsg] {
						return false, desc
					}
				}
			}
			return true, desc
		},
		func(states map[int]State, leaf bool) (bool, string) {
			// RB4 Agreement
			desc := "RB4: Agreement"
			// Use leaf nodes to check for liveness properties
			// Can not say that the predicate has been broken for non-leaf nodes
			if !leaf {
				return true, desc
			}
			delivered := map[message]bool{}
			for _, node := range states {
				for msg := range node.delivered {
					delivered[msg] = true
				}
			}
			for msg := range delivered {
				for _, node := range states {
					if !node.delivered[msg] {
						return false, desc
					}
				}
			}
			return true, desc
		},
	)

	if resp := checker.Check(&sm.StateRoot); !resp.Result {
		fmt.Println("Node broke predicate:", resp.Test)
		for _, state := range resp.Sequence {
			fmt.Println(state)
		}
	}
}
