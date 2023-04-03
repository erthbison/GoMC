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

	"gomc/examples/paxos"
	"gomc/runner"
	"gomc/runner/controller"
	"gomc/runner/recorder"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addrMap = map[int64]string{
		1: "127.0.0.1:50000",
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
	addr2id := map[string]int{}
	nodeIds := []int{}
	for id, addr := range addrMap {
		nodeIds = append(nodeIds, int(id))
		addr2id[addr] = int(id)
		tmp[int(id)] = addr
	}

	gnc := controller.NewGrpcNodeController(addr2id)
	r := runner.NewRunner(time.Second, gnc, gnc, func(t *paxos.Server) error { t.Stop(); return nil })
	r.Start(
		func() map[int]*paxos.Server {
			nodes := make(map[int]*paxos.Server)
			for id := range addrMap {
				srv, err := paxos.NewServer(id, addrMap, func(int) {}, grpc.UnaryInterceptor(gnc.ServerInterceptor(int(id))))
				r.CrashSubscribe(srv.NodeCrash)
				if err != nil {
					panic(err)
				}
				nodes[int(id)] = srv
			}
			return nodes
		},
		tmp,
		func(srv *paxos.Server, lis net.Listener, dial runner.Dialer) {
			go srv.StartServer(lis)

			err := srv.DialNodes(
				grpc.WithContextDialer(func(_ context.Context, addr string) (net.Conn, error) { return dial(addr) }),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithUnaryInterceptor(gnc.ClientInterceptor(int(srv.Id))),
			)
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
		chn := r.SubscribeStateUpdates()
		for state := range chn {
			// fmt.Println(state)
			fmt.Fprintln(f, state)
		}
	}()

	go func(c <-chan recorder.Message) {
		messages := map[int][]recorder.Message{}
		for _, id := range addr2id {
			messages[id] = make([]recorder.Message, 0)
		}
		m, err := os.Create("MESSAGES.txt")
		if err != nil {
			panic(err)
		}
		for msg := range c {
			var id int
			if msg.Sent {
				id = msg.From
			} else {
				id = msg.To
			}
			m := messages[id]
			m = append(m, msg)
			messages[id] = m
		}

		for id, msg := range messages {
			fmt.Fprintf(m, "Node %v: \n%v \n", id, msg)
		}
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
			gnc.Stop()
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
	time.Sleep(time.Second)
}
