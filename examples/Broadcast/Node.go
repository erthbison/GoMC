package main

import "time"

type DeliverMsg struct {
	From    int
	To      int
	Message []byte
}

type AckMsg struct {
	From    int
	To      int
	Message []byte
}

type Node struct {
	Id        int
	nodes     []int
	Delivered int
	Acked     int
	send      func(int, int, string, []byte)
	timeout   func(time.Duration) <-chan time.Time
}

func (n *Node) Broadcast(message []byte) {
	<-n.timeout(time.Second)
	for _, id := range n.nodes {
		n.send(
			n.Id,
			id,
			"Deliver",
			message,
		)
	}
}

func (n *Node) Deliver(from, to int, message []byte) {
	n.Delivered++
	for _, id := range n.nodes {
		n.send(
			n.Id,
			id,
			"Ack",
			message,
		)
	}
}

func (n *Node) Ack(from, to int, message []byte) {
	n.Acked++
}
