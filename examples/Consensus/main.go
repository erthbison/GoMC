package main

import (
	"fmt"
	"gomc/examples/server"
	"gomc/runner"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc/test/bufconn"
)

func init() {
	server.RegisterTypes(Value[int]{})
	server.RegisterTypes(Value[string]{})
}

func main() {
	runRunner()
}

func runServer() {
	addrs := map[int]string{
		1: ":50000",
		2: ":50001",
		3: ":50002",
		4: ":50003",
		5: ":50004",
	}

	nodeIds := []uint{}
	listeners := map[string]*bufconn.Listener{}
	for id, addr := range addrs {
		nodeIds = append(nodeIds, uint(id))
		lis := bufconn.Listen(1024)
		listeners[addr] = lis
	}

	dial := func(s string) (net.Conn, error) {
		return listeners[s].Dial()
	}
	var wait sync.WaitGroup
	wait.Add(len(addrs))
	for id, addr := range addrs {
		server := server.NewServer()
		n := NewHierarchicalConsensus[int](
			uint(id),
			nodeIds,
			server.SendFactory("hc"),
		)

		err := server.RegisterModule("hc", n)
		if err != nil {
			panic(err)
		}
		server.Subscribe("Crash", "hc")
		server.Start(listeners[addr])

		go func(i int, node *HierarchicalConsensus[int]) {
			server.DialNodes(addrs, dial)
			val := Value[int]{i}
			node.Propose(val)
			fmt.Printf("Node %v proposed %v\n", node.id, val)
			fmt.Printf("Node %v Decided %v\n", node.id, <-n.DecidedSignal)
			wait.Done()
		}(id, n)
	}
	wait.Wait()
}

type State struct {
	proposed Value[string]
	decided  []Value[string]
}

type Node struct {
	srv   *server.Server
	c     *HierarchicalConsensus[string]
	addrs map[int]string
}

func NewNode(id int, nodeIds []uint, addrs map[int]string) *Node {
	server := server.NewServer()
	n := NewHierarchicalConsensus[string](
		uint(id),
		nodeIds,
		server.SendFactory("hc"),
	)
	return &Node{
		srv:   server,
		c:     n,
		addrs: addrs,
	}
}

func (n *Node) Start(lis net.Listener, dial func(string) (net.Conn, error)) {
	err := n.srv.RegisterModule("hc", n.c)
	if err != nil {
		panic(err)
	}
	n.srv.Subscribe("Crash", "hc")
	n.srv.Start(lis)
	n.srv.DialNodes(n.addrs, dial)
}

func (n *Node) State() any {
	decided := make([]Value[string], len(n.c.DecidedVal))
	copy(decided, n.c.DecidedVal)
	return State{
		proposed: n.c.ProposedVal,
		decided:  decided,
	}
}

func (n *Node) Propose(val Value[string]) {
	n.c.Propose(val)
}

func (n *Node) Pause() error {
	return n.srv.Pause()
}

func (n *Node) Resume() error {
	return n.srv.Resume()
}

func runRunner() {
	addrs := map[int]string{
		1: ":50000",
		2: ":50001",
		3: ":50002",
		4: ":50003",
		5: ":50004",
	}

	nodeIds := []uint{}
	for id := range addrs {
		nodeIds = append(nodeIds, uint(id))
	}

	r := runner.NewRunner(time.Second)
	go r.Start(func(id int) runner.Node {
		return NewNode(id, nodeIds, addrs)
	}, addrs)

	go func() {
		chn := r.GetStateUpdates()
		for state := range chn {
			fmt.Println(state)
		}
	}()

	<-time.After(5 * time.Second)
	go func() {
		r.PauseNode(3)
		<-time.After(5 * time.Second)
		r.ResumeNode(3)
	}()
	for _, id := range nodeIds {
		r.Request(int(id), "Propose", Value[string]{Val: addrs[int(id)]})
	}
	fmt.Scanln()
	r.Stop()
	fmt.Scanln()
}
