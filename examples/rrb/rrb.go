package main

import (
	"fmt"
)

type message struct {
	From    int
	Index   int
	Payload string
}

func (m message) String() string {
	return fmt.Sprintf("{F:%v, i:%v}", m.From, m.Index)
}

type Rrb struct {
	id    int
	nodes []int

	delivered map[message]bool
	sent      map[message]bool
	send      func(to int, msgType string, msg ...any)

	crashed bool

	deliveredSlice []message
}

func NewRrb(id int, nodes []int, send func(int, string, ...any)) *Rrb {
	return &Rrb{
		id:    id,
		nodes: nodes,

		delivered: make(map[message]bool),
		sent:      make(map[message]bool),
		send:      send,
	}
}

func (rrb *Rrb) Broadcast(msg string) {
	if rrb.crashed {
		return
	}
	message := message{
		From:    rrb.id,
		Index:   len(rrb.sent),
		Payload: msg,
	}

	for _, target := range rrb.nodes {
		rrb.send(target, "Deliver", message)
	}
	rrb.sent[message] = true
}

func (rrb *Rrb) Deliver(message message) {
	if rrb.crashed {
		return
	}

	// violation of RB2:No Duplication
	rrb.deliveredSlice = append(rrb.deliveredSlice, message)
	if !rrb.delivered[message] {
		rrb.delivered[message] = true
		for _, target := range rrb.nodes {
			rrb.send(target, "Deliver", message)
		}
	}
}
