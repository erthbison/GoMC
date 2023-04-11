package main

import (
	"context"
	"fmt"
	"gomc/examples/paxos"
	"log"
	"net"
	"sync"

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
		4: ":50003",
		5: ":50004",
	}
	addrToIdMao = map[string]int{}
	lisMap      = map[string]*bufconn.Listener{}
	nodes       = map[int64]*paxos.Server{}

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
		srv, err := paxos.NewServer(id, addrMap, func(int) {})
		if err != nil {
			log.Panicln(err)
		}
		go func(addr string) {
			err := srv.StartServer(lisMap[addr])
			if err != nil {
				log.Println(err)
			}
		}(addr)
		nodes[id] = srv
	}

	for _, srv := range nodes {
		srv.DialNodes(
			grpc.WithContextDialer(dial),
			grpc.WithBlock(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
	}

	// fmt.Scanln()
	var wait sync.WaitGroup
	for _, node := range nodes {
		wait.Add(1)
		go func(n *paxos.Server) {
			val, ok := <-n.Subscribe()
			if !ok {
				fmt.Println("Node", n.Id, "Crashed")
			} else {
				fmt.Println("Node", n.Id, "Decided: ", val)
			}
			wait.Done()
		}(node)
	}

	nodes[5].Stop()
	go func() {
		for _, n := range nodes {
			n.NodeCrash(5, false)
		}
	}()

	for _, node := range nodes {
		go func(n *paxos.Server) {
			fmt.Println("Node", n.Id, "Propose", addrMap[n.Id])
			n.Propose(addrMap[n.Id])
		}(node)
	}
	wait.Wait()
}
