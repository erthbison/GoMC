package paxos

import (
	"bytes"
	"context"
	"encoding/json"
	"gomc"
	"gomc/gomcGrpc"
	"gomc/predicate"
	"net"
	"os"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
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

func TestPaxosSim(t *testing.T) {
	for id, addr := range addrMap {
		addrToIdMap[addr] = int(id)
	}

	sim := gomc.Prepare[Server, State](
		gomc.RandomWalkScheduler(1),
		gomc.MaxDepth(100000),
		gomc.MaxRuns(1000),
	)
	w, err := os.Create("export.txt")
	if err != nil {
		t.Errorf("Error while creating file: %v", err)
	}
	defer w.Close()
	resp := sim.RunSimulation(
		gomc.InitNodeFunc(func(sp gomc.SimulationParameters) map[int]*Server {
			lisMap := map[string]*bufconn.Listener{}
			for _, addr := range addrMap {
				lisMap[addr] = bufconn.Listen(bufSize)
			}
			gem := gomcGrpc.NewGrpcEventManager(addrToIdMap, sp.Sch, sp.NextEvt)

			nodes := make(map[int]*Server)
			for id, addr := range addrMap {
				srv, err := NewServer(id, addrMap, gem.WaitForSend)
				if err != nil {
					t.Errorf("Error while starting simulation: %v", err)
				}
				go srv.StartServer(lisMap[addr])
				sp.Fm.Subscribe(srv.NodeCrash)
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
		gomc.WithTreeStateManager(
			func(t *Server) State {
				return State{
					proposed: t.proposal,
					decided:  t.Learner.val.GetVal(),
				}
			},
			func(s1, s2 State) bool {
				return s1 == s2
			},
		),
		gomc.WithPredicate(
			func(gs gomc.GlobalState[State], terminal bool, seq []gomc.GlobalState[State]) bool {
				// Only a value that has been proposed may be chosen
				proposedVal := map[string]bool{}
				for _, state := range gs.LocalStates {
					proposedVal[state.proposed] = true
				}
				return predicate.ForAllNodes(func(s State) bool { return !(s.decided != zeroVal && !proposedVal[s.decided]) }, gs, false)
			},
			func(gs gomc.GlobalState[State], terminal bool, seq []gomc.GlobalState[State]) bool {
				// Only a single value is chosen
				decidedVal := map[string]bool{}
				for _, state := range gs.LocalStates {
					if state.decided != zeroVal {
						decidedVal[state.decided] = true
					}
				}
				if len(decidedVal) > 1 {
					return false
				}
				return true
			},
			predicate.Eventually(
				func(gs gomc.GlobalState[State], seq []gomc.GlobalState[State]) bool {
					// All correct node should eventually learn the decided value
					return predicate.ForAllNodes(func(s State) bool { return s.decided != zeroVal }, gs, true)
				},
			),
		),
		gomc.IncorrectNodes(func(t *Server) { t.Stop() }, 5, 1),
		gomc.Export(w),
	)
	if ok, text := resp.Response(); !ok {
		t.Errorf("Test Failed: \n %v", text)

		var buffer bytes.Buffer
		json.NewEncoder(&buffer).Encode(resp.Export())
		os.WriteFile("FailedRun.txt", buffer.Bytes(), 0755)
	}
}

func TestPaxosReplay(t *testing.T) {
	for id, addr := range addrMap {
		addrToIdMap[addr] = int(id)
	}

	in, err := os.ReadFile("FailedRun.txt")
	if err != nil {
		t.Errorf("Error while setting up test: %v", err)
	}
	buffer := bytes.NewBuffer(in)
	var run []string
	json.NewDecoder(buffer).Decode(&run)

	sim := gomc.Prepare[Server, State](gomc.ReplayScheduler(run), gomc.MaxDepth(100000))
	// sim := gomc.Prepare[Server, State](gomc.WithScheduler(scheduler.NewGuidedSearch(scheduler.NewRandomScheduler(25, 1), run)))
	w, err := os.Create("export.txt")
	if err != nil {
		t.Errorf("Error while creating file: %v", err)
	}
	defer w.Close()
	resp := sim.RunSimulation(
		gomc.InitNodeFunc(func(sp gomc.SimulationParameters) map[int]*Server {
			lisMap := map[string]*bufconn.Listener{}
			for _, addr := range addrMap {
				lisMap[addr] = bufconn.Listen(bufSize)
			}
			gem := gomcGrpc.NewGrpcEventManager(addrToIdMap, sp.Sch, sp.NextEvt)

			nodes := make(map[int]*Server)
			for id, addr := range addrMap {
				srv, err := NewServer(id, addrMap, gem.WaitForSend)
				if err != nil {
					t.Errorf("Error while starting simulation: %v", err)
				}
				go srv.StartServer(lisMap[addr])
				sp.Fm.Subscribe(srv.NodeCrash)
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
		gomc.WithTreeStateManager(
			func(t *Server) State {
				return State{
					proposed: t.proposal,
					decided:  t.Learner.val.GetVal(),
				}
			},
			func(s1, s2 State) bool {
				return s1 == s2
			},
		),
		gomc.WithPredicate(
			func(gs gomc.GlobalState[State], terminal bool, seq []gomc.GlobalState[State]) bool {
				// Only a value that has been proposed may be chosen
				proposedVal := map[string]bool{}
				for _, state := range gs.LocalStates {
					proposedVal[state.proposed] = true
				}
				for _, state := range gs.LocalStates {
					if state.decided != zeroVal && !proposedVal[state.decided] {
						return false
					}
				}
				return true
			},
			func(gs gomc.GlobalState[State], terminal bool, seq []gomc.GlobalState[State]) bool {
				// Only a single value is chosen
				decidedVal := map[string]bool{}
				for _, state := range gs.LocalStates {
					if state.decided != zeroVal {
						decidedVal[state.decided] = true
					}
				}
				if len(decidedVal) > 1 {
					return false
				}
				return true
			},
			predicate.Eventually(
				func(gs gomc.GlobalState[State], seq []gomc.GlobalState[State]) bool {
					// All correct node should eventually learn the decided value
					for id, state := range gs.LocalStates {
						if gs.Correct[id] {
							if state.decided == zeroVal {
								return false
							}
						}
					}
					return true
				},
			),
		),
		gomc.IncorrectNodes(func(t *Server) { t.Stop() }, 5, 1),
		gomc.Export(w),
	)
	if ok, text := resp.Response(); !ok {
		t.Errorf("Test Failed: \n %v", text)
	}
}