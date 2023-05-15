package paxos

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"os"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"gomc"
	"gomc/checking"
	"gomc/eventManager"
)

type State struct {
	proposed string
	decided  string
}

var (
	bufSize = 1024
	addrMap = map[int64]string{
		1: ":50000",
		2: ":50001",
		3: ":50002",
		4: ":50003",
		5: ":50004",
	}
	addrToIdMap = map[string]int{}

	zeroVal string
)

var predicates = []checking.Predicate[State]{
	func(s checking.State[State]) bool {
		// Only a value that has been proposed may be chosen
		proposedVal := map[string]bool{}
		for _, state := range s.LocalStates {
			proposedVal[state.proposed] = true
		}
		return checking.ForAllNodes(func(s State) bool { return !(s.decided != zeroVal && !proposedVal[s.decided]) }, s, false)
	},
	func(s checking.State[State]) bool {
		// Only a single value is chosen
		decidedVal := map[string]bool{}
		for _, state := range s.LocalStates {
			if state.decided != zeroVal {
				decidedVal[state.decided] = true
			}
		}
		if len(decidedVal) > 1 {
			return false
		}
		return true
	},
	checking.Eventually(
		func(s checking.State[State]) bool {
			// All correct node should eventually learn the decided value
			return checking.ForAllNodes(func(s State) bool { return s.decided != zeroVal }, s, true)
		},
	),
}

func TestPaxosSim(t *testing.T) {
	for id, addr := range addrMap {
		addrToIdMap[addr] = int(id)
	}

	sim := gomc.PrepareSimulation(
		gomc.WithTreeStateManager(
			func(t *Server) State {
				return State{
					proposed: t.Proposal,
					decided:  t.Decided,
				}
			},
			func(s1, s2 State) bool {
				return s1 == s2
			},
		),
		gomc.RandomWalkScheduler(1),
	)
	w, err := os.Create("export.txt")
	if err != nil {
		t.Errorf("Error while creating file: %v", err)
	}
	defer w.Close()
	resp := sim.Run(
		gomc.InitNodeFunc(func(sp eventManager.SimulationParameters) map[int]*Server {
			lisMap := map[string]*bufconn.Listener{}
			for _, addr := range addrMap {
				lisMap[addr] = bufconn.Listen(bufSize)
			}
			gem := eventManager.NewGrpcEventManager(addrToIdMap, sp.EventAdder, sp.NextEvt)

			nodes := make(map[int]*Server)
			for id, addr := range addrMap {
				srv, err := NewServer(id, addrMap, gem.WaitForSend(int(id)))
				if err != nil {
					t.Errorf("Error while starting simulation: %v", err)
				}
				go srv.StartServer(lisMap[addr])
				sp.CrashSubscribe(int(id), srv.NodeCrash)
				nodes[int(id)] = srv
			}

			for id, node := range nodes {
				node.DialNodes(
					grpc.WithUnaryInterceptor(gem.UnaryClientControllerInterceptor(id)),
					grpc.WithContextDialer(
						func(ctx context.Context, s string) (net.Conn, error) {
							return lisMap[s].DialContext(ctx)
						},
					),
					grpc.WithBlock(),
					grpc.WithTransportCredentials(insecure.NewCredentials()),
				)
			}
			return nodes
		}),
		gomc.WithRequests(
			gomc.NewRequest(1, "Propose", "1"),
			gomc.NewRequest(2, "Propose", "2"),
			gomc.NewRequest(3, "Propose", "3"),
			gomc.NewRequest(4, "Propose", "4"),
			gomc.NewRequest(5, "Propose", "5"),
		),
		gomc.WithPredicateChecker(predicates...),
		gomc.WithPerfectFailureManager(func(t *Server) { t.Stop() }, 5, 1),
		gomc.WithStopFunctionSimulator(func(t *Server) { t.Stop() }),
		gomc.Export(w),
	)
	if ok, text := resp.Response(); !ok {
		t.Errorf("Test Failed: \n %v", text)

		var buffer bytes.Buffer
		json.NewEncoder(&buffer).Encode(resp.Export())
		os.WriteFile("FailedRun.txt", buffer.Bytes(), 0755)
	}
}
