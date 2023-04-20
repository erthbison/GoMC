package main

import (
	"context"
	"fmt"
	"gomc/examples/multipaxos"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

var (
	bufSize = 1024
	addrMap = map[int64]string{
		1: ":50000",
		2: ":50001",
		3: ":50002",
		// 4: ":50003",
		// 5: ":50004",
	}
	addrToIdMao = map[string]int{}
	lisMap      = map[string]*bufconn.Listener{}
	nodes       = map[int64]*multipaxos.MultiPaxos{}

	dial = func(ctx context.Context, s string) (net.Conn, error) {
		return lisMap[s].DialContext(ctx)
	}
)

func main() {

	for id, addr := range addrMap {
		addrToIdMao[addr] = int(id)
		lisMap[addr] = bufconn.Listen(bufSize)
	}

	for id, addr := range addrMap {
		mp := multipaxos.NewMultiPaxos(id, addrMap, func(i int) {})
		mp.Start(lisMap[addr])
		nodes[id] = mp
	}

	for _, mp := range nodes {
		mp.DialNodes(
			grpc.WithContextDialer(dial),
			grpc.WithBlock(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
	}

	fmt.Scanln()
	// var wait sync.WaitGroup
	for id, node := range nodes {
		func(id int64) {
			node.LearnSubscribe(func(s string, i int64) {
				fmt.Printf("Node: %v Decided In Slot %v Val: %v\n", id, i, s)
			})
		}(id)
	}

	nodes[3].ProposeVal("TEST")
	fmt.Scanln()
	nodes[3].ProposeVal("TEST2")
	nodes[3].Stop()
	for _, n := range nodes {
		n.NodeCrash(3, false)
	}

	fmt.Scanln()

	// time.Sleep(1 * time.Second)
	nodes[2].ProposeVal("TEST3")
	fmt.Scanln()

	// time.Sleep(5 * time.Second)
	// fmt.Println("DONE")

	for id, n := range nodes {
		fmt.Println("Node", id, "Learnt", n.LearntValues)
	}
}
