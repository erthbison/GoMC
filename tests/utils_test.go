package gomc_test

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

type BroadcastNode struct {
	Id        int
	nodes     []int
	Delivered int
	Acked     int
	send      func(int, string, ...any)
}

func (n *BroadcastNode) Broadcast(message []byte) {
	for _, id := range n.nodes {
		n.send(
			id,
			"Deliver",
			message,
		)
	}
}

func (n *BroadcastNode) Deliver(message []byte) {
	n.Delivered++
	for _, id := range n.nodes {
		n.send(
			id,
			"Ack",
			message,
		)
	}
}

func (n *BroadcastNode) Ack(message []byte) {
	n.Acked++
}

type BroadcastState struct {
	delivered int
	acked     int
}
