package main

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const (
	bufSize = 1024
)

func main() {
	addrMap := map[int32]string{
		1: ":50000",
		2: ":50001",
		3: ":50002",
		4: ":50003",
		5: ":50004",
	}
	ids := []int32{1, 2, 3, 4, 5}
	addrToIdMao := map[string]int{}
	for id, addr := range addrMap {
		addrToIdMao[addr] = int(id)
	}

	lisMap := map[string]*bufconn.Listener{}
	for _, addr := range addrMap {
		lisMap[addr] = bufconn.Listen(bufSize)
	}

	nodes := map[int32]*GrpcConsensus{}
	for id, addr := range addrMap {
		gc := NewGrpcConsensus(id, lisMap[addr], func(int, int) {})
		nodes[id] = gc
	}

	for _, node := range nodes {
		node.DialServers(
			ids,
			addrMap,
			grpc.WithContextDialer(
				func(ctx context.Context, s string) (net.Conn, error) {
					return lisMap[s].DialContext(ctx)
				},
			),
			grpc.WithBlock(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
	}

	fmt.Scanln()

	nodes[2].Stop()
	for _, node := range nodes {
		go func(n *GrpcConsensus) {
			n.Crash(2)
			n.Propose(addrMap[n.id])
		}(node)
	}
	fmt.Scanln()
}
