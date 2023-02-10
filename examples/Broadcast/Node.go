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
	send      func(int, string, []byte)
	sleep     func(time.Duration)
}

func (n *Node) Broadcast(message []byte) {
	n.sleep(time.Second)
	for _, id := range n.nodes {
		n.send(
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
			id,
			"Ack",
			message,
		)
	}
}

func (n *Node) Ack(from, to int, message []byte) {
	n.Acked++
}
