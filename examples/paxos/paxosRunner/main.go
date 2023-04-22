package main

import (
	"bufio"
	"context"
	"fmt"
	"gomc"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"gomc/examples/paxos"
	"gomc/gomcGrpc"
	"gomc/runner"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

var (
	bufSize = 1024
	addrMap = map[int64]string{
		1: "127.0.0.1:50000",
		2: ":50001",
		3: ":50002",
		4: ":50003",
		5: ":50004",
	}
	addr2id = map[string]int{}
)

type State struct {
	proposed string
	decided  string
}

func main() {

	for id, addr := range addrMap {
		addr2id[addr] = int(id)
	}
	r := gomc.Run(
		gomc.InitNodeFunc(
			func(sp gomc.SimulationParameters) map[int]*paxos.Server {
				lisMap := map[string]*bufconn.Listener{}
				for _, addr := range addrMap {
					lisMap[addr] = bufconn.Listen(bufSize)
				}

				gem := gomcGrpc.NewGrpcEventManager(addr2id, sp.EventAdder, sp.NextEvt)
				nodes := make(map[int]*paxos.Server)
				for id, addr := range addrMap {
					srv, err := paxos.NewServer(id, addrMap, gem.WaitForSend(int(id)))
					sp.CrashSubscribe(srv.NodeCrash)
					if err != nil {
						panic(err)
					}
					nodes[int(id)] = srv
					go srv.StartServer(lisMap[addr])
				}

				for id, srv := range nodes {
					err := srv.DialNodes(
						grpc.WithContextDialer(
							func(ctx context.Context, s string) (net.Conn, error) {
								return lisMap[s].DialContext(ctx)
							},
						),
						grpc.WithTransportCredentials(insecure.NewCredentials()),
						grpc.WithUnaryInterceptor(gem.UnaryClientControllerInterceptor(id)),
					)
					if err != nil {
						panic(err)
					}
				}
				return nodes
			},
		),
		gomc.WithStateFunction(func(t *paxos.Server) State {
			t.Lock()
			defer t.Unlock()
			return State{
				proposed: t.Proposal,
				decided:  t.Decided,
			}
		}),
	)

	wait := new(sync.WaitGroup)
	wait.Add(1)

	// Store records by node id and write them to file
	go func(c <-chan runner.Record) {
		messages := map[int][]runner.Record{}
		for _, id := range addr2id {
			messages[id] = make([]runner.Record, 0)
		}
		m, err := os.Create("Messages.txt")
		if err != nil {
			panic(err)
		}
		for rec := range c {
			m := messages[rec.Target()]
			m = append(m, rec)
			messages[rec.Target()] = m
		}

		for id, msgSlice := range messages {
			fmt.Fprintf(m, "Node %v: [\n", id)
			for _, msg := range msgSlice {
				fmt.Fprintf(m, "\t%v \n", msg)
			}
			fmt.Fprint(m, "]\n")
		}
		wait.Done()
	}(r.SubscribeMessages())

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
			err = r.Request(id, "Propose", params[1])
			if err != nil {
				panic(fmt.Errorf("Error: %v", err))
			}
		case "pause":
			id, err := strconv.Atoi(params[0])
			if err != nil {
				panic(fmt.Errorf("Error: %v", err))
			}
			err = r.PauseNode(id)
			if err != nil {
				panic(fmt.Errorf("Error: %v", err))
			}
		case "resume":
			id, err := strconv.Atoi(params[0])
			if err != nil {
				panic(fmt.Errorf("Error: %v", err))
			}
			err = r.ResumeNode(id)
			if err != nil {
				panic(fmt.Errorf("Error: %v", err))
			}
		case "crash":
			id, err := strconv.Atoi(params[0])
			if err != nil {
				panic(fmt.Errorf("Error: %v", err))
			}
			err = r.CrashNode(id)
			if err != nil {
				panic(fmt.Errorf("Error: %v", err))
			}
		default:
			fmt.Println("Invalid command")
		}
	}

	// r.Request(5, "Propose", "test")

	wait.Wait()
}
