package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"gomc"
	"gomc/examples/paxos"
	"gomc/runnerControllers"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addrMap = map[int64]string{
		1: ":50000",
		2: ":50001",
		3: ":50002",
		4: ":50003",
		5: ":50004",
	}
)

type State struct {
	proposed string
	decided  string
}

func main() {
	tmp := map[int]string{}
	nodeIds := []int{}
	for id, addr := range addrMap {
		nodeIds = append(nodeIds, int(id))
		tmp[int(id)] = addr
	}

	gnc := runnerControllers.NewGrpcNodeController(nodeIds)
	r := gomc.NewRunner(time.Second, gnc, func(t *paxos.Server) error { t.Stop(); return nil })
	go r.Start(
		func() map[int]*paxos.Server {
			nodes := make(map[int]*paxos.Server)
			for id := range addrMap {
				srv, err := paxos.NewServer(id, addrMap, func(i1, i2 int) {}, grpc.UnaryInterceptor(gnc.ServerInterceptor(int(id))))
				r.CrashSubscribe(srv.NodeCrash)
				if err != nil {
					panic(err)
				}
				nodes[int(id)] = srv
			}
			return nodes
		},
		tmp,
		func(srv *paxos.Server, lis net.Listener, dial gomc.Dialer) {
			go srv.StartServer(lis)

			err := srv.DialNodes(grpc.WithContextDialer(func(_ context.Context, addr string) (net.Conn, error) { return dial(addr) }), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				panic(err)
			}
		},
		func(t *paxos.Server) any {
			t.Lock.Lock()
			defer t.Lock.Unlock()
			return State{
				proposed: t.Proposal,
				decided:  t.Learner.Val.GetVal(),
			}
		},
	)

	go func() {
		f, err := os.Create("TMP.txt")
		if err != nil {
			panic(err)
		}
		chn := r.GetStateUpdates()
		for state := range chn {
			// fmt.Println(state)
			fmt.Fprintln(f, state)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	ok := true
	for ok {
		scanner.Scan()
		vals := strings.Split(scanner.Text(), " ")
		cmd := vals[0]
		params := vals[1:]

		switch cmd {
		case "stop":
			r.Stop()
			ok = false
		case "propose":
			if len(params) < 2 {
				panic("To few parameters")
			}
			id, err := strconv.Atoi(params[0])
			if err != nil {
				panic(fmt.Errorf("Error: %v", err))
			}
			r.Request(id, "Propose", params[1])
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
