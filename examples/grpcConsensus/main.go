package main

import (
	"context"
	"fmt"
	"net"
	"strconv"

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
	}

	lisMap := map[string]*bufconn.Listener{}
	for _, addr := range addrMap {
		lisMap[addr] = bufconn.Listen(bufSize)
	}

	nodes := map[int32]*GrpcConsensus{}
	for id, addr := range addrMap {
		gc := NewGrpcConsensus(id, lisMap[addr])
		nodes[id] = gc
	}

	for _, node := range nodes {
		node.DialServers(
			addrMap,
			grpc.WithContextDialer(
				func(ctx context.Context, s string) (net.Conn, error) {
					return lisMap[s].DialContext(ctx)
				},
			),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
	}

	for id, node := range nodes {
		node.Propose(strconv.Itoa(int(id)))
	}
	fmt.Scanln()
}
