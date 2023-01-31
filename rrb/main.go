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
	tester := tester.CreateTester[Rrb, State](sch, sm)
	tester.Simulate(
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
					tester.Send,
				)
			}
			return nodes
		},
		func(nodes map[int]*Rrb) {
			for i := 0; i < 1; i++ {
				nodes[0].Broadcast(fmt.Sprintf("Test Message - %v", i))
				// message := message{
				// 	From:    0,
				// 	Index:   0,
				// 	Payload: "test",
				// }
				// byteMsg, err := json.Marshal(message)
				// if err != nil {
				// 	panic(err)
				// }
				// nodes[1].Deliver(0, 1, byteMsg)
			}
		},
	)
	fmt.Println(sch.EventRoot)
	fmt.Println(sm.StateRoot)

	fmt.Println(sm.StateRoot.Newick())
	fmt.Println(sch.EventRoot.Newick())
}
