package main

import (
	"bufio"
	"fmt"
	"gomc"
	"gomc/examples/server"
	"net"
	"os"
	"strconv"
	"strings"
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

	nodeIds := []int{}
	listeners := map[string]*bufconn.Listener{}
	for id, addr := range addrs {
		nodeIds = append(nodeIds, id)
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
			id,
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

func NewNode(id int, nodeIds []int, addrs map[int]string) *Node {
	server := server.NewServer()
	n := NewHierarchicalConsensus[string](
		id,
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

type ServerController struct {
	nodes map[int]*Node
}

func (sc *ServerController) Pause(id int) error {
	return sc.nodes[id].Pause()
}

func (sc *ServerController) Resume(id int) error {
	return sc.nodes[id].Resume()
}

func runRunner() {
	addrs := map[int]string{
		1: ":50000",
		2: ":50001",
		3: ":50002",
		4: ":50003",
		5: ":50004",
	}

	nodeIds := []int{}
	for id := range addrs {
		nodeIds = append(nodeIds, id)
	}
	sc := &ServerController{
		nodes: make(map[int]*Node),
	}

	r := gomc.NewRunner(time.Second, sc, func(t *Node) error { t.c.crashed = true; return nil })
	go r.Start(
		func() map[int]*Node {
			nodes := map[int]*Node{}
			for id := range addrs {
				n := NewNode(id, nodeIds, addrs)
				r.CrashSubscribe(n.c.Crash)
				nodes[id] = n
				sc.nodes[id] = n
			}
			return nodes
		},
		addrs,
		func(n *Node, lis net.Listener, dial gomc.Dialer) {
			n.Start(lis, dial)
		},
		func(t *Node) any {
			return t.State()
		},
	)

	go func() {
		f, err := os.Create("TMP.txt")
		if err != nil {
			panic(err)
		}
		chn := r.GetStateUpdates()
		for state := range chn {
			fmt.Fprintln(f, state)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	var done bool
	for !done {
		scanner.Scan()
		vals := strings.Split(scanner.Text(), " ")
		cmd := vals[0]
		params := vals[1:]

		switch cmd {
		case "stop":
			r.Stop()
			done = false
		case "propose":
			if len(params) < 2 {
				panic("To few parameters")
			}
			id, err := strconv.Atoi(params[0])
			if err != nil {
				panic(fmt.Errorf("Error: %v", err))
			}
			val := Value[string]{Val: params[1]}
			r.Request(id, "Propose", val)
		case "pause":
			id, err := strconv.Atoi(params[0])
			if err != nil {
				panic(fmt.Errorf("Error: %v", err))
			}
			err = r.PauseNode(id)
			if err != nil {
				fmt.Println("Received error:", err)
			}
		case "resume":
			id, err := strconv.Atoi(params[0])
			if err != nil {
				panic(fmt.Errorf("Error: %v", err))
			}
			err = r.ResumeNode(id)
			if err != nil {
				fmt.Println("Received error:", err)
			}
		case "crash":
			id, err := strconv.Atoi(params[0])
			if err != nil {
				panic(fmt.Errorf("Error: %v", err))
			}
			err = r.CrashNode(id)
			if err != nil {
				fmt.Println("Received error:", err)
			}
		default:
			fmt.Println("Invalid command")
		}
	}
}
