package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"testing"

	"gomc"
	"gomc/gomcGrpc"
	"gomc/predicate"

	"golang.org/x/exp/slices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

var predicates = []gomc.Predicate[state]{
	predicate.Eventually(
		// C1: Termination
		func(gs gomc.GlobalState[state], _ []gomc.GlobalState[state]) bool {
			return predicate.ForAllNodes(func(s state) bool {
				return len(s.decided) > 0
			}, gs, true)
		},
	),
	func(gs gomc.GlobalState[state], _ bool, _ []gomc.GlobalState[state]) bool {
		// C2: Validity
		proposed := make(map[string]bool)
		for _, node := range gs.LocalStates {
			proposed[node.proposed] = true
		}
		return predicate.ForAllNodes(func(s state) bool {
			if len(s.decided) < 1 {
				// The process has not decided a value yet
				return true
			}
			return proposed[s.decided[0]]
		}, gs, false)
	},
	func(gs gomc.GlobalState[state], _ bool, seq []gomc.GlobalState[state]) bool {
		// C3: Integrity
		return predicate.ForAllNodes(func(s state) bool { return len(s.decided) < 2 }, gs, false)
	},
	func(gs gomc.GlobalState[state], _ bool, seq []gomc.GlobalState[state]) bool {
		// C4: Agreement
		decided := make(map[string]bool)
		predicate.ForAllNodes(func(s state) bool {
			for _, val := range s.decided {
				decided[val] = true
			}
			return true
		}, gs, true)
		return len(decided) <= 1
	},
}

func TestGrpcConsensus(t *testing.T) {
	sim := gomc.Prepare[GrpcConsensus, state](
		gomc.RandomWalkScheduler(1),
		gomc.MaxRuns(1000),
	)

	sm := gomc.NewTreeStateManager(
		func(node *GrpcConsensus) state {
			decided := make([]string, len(node.DecidedVal))
			copy(decided, node.DecidedVal)
			return state{
				proposed: node.ProposedVal,
				decided:  decided,
			}
		},
		func(a, b state) bool {
			if a.proposed != b.proposed {
				return false
			}
			return slices.Equal(a.decided, b.decided)
		},
	)
	ids := []int32{1, 2, 3, 4, 5}
	addrMap := map[int32]string{
		1: ":50000",
		2: ":50001",
		3: ":50002",
		4: ":50003",
		5: ":50004",
	}

	addrToIdMap := map[string]int{}
	for id, addr := range addrMap {
		addrToIdMap[addr] = int(id)
	}

	resp := sim.RunSimulation(
		gomc.InitNodeFunc(
			func(sp gomc.SimulationParameters) map[int]*GrpcConsensus {
				gem := gomcGrpc.NewGrpcEventManager(addrToIdMap, sp.Sch, sp.NextEvt)
				lisMap := map[string]*bufconn.Listener{}
				for _, addr := range addrMap {
					lisMap[addr] = bufconn.Listen(bufSize)
				}

				nodes := map[int]*GrpcConsensus{}
				for id, addr := range addrMap {
					gc := NewGrpcConsensus(id, lisMap[addr], gem.WaitForSend)
					sp.Fm.Subscribe(gc.Crash)
					nodes[int(id)] = gc
				}

				for id, node := range nodes {
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
						grpc.WithUnaryInterceptor(gem.UnaryClientControllerInterceptor(int(id))),
					)
				}
				return nodes
			},
		),
		gomc.WithRequests(
			gomc.NewRequest(1, "Propose", "2"),
			gomc.NewRequest(2, "Propose", "3"),
			gomc.NewRequest(3, "Propose", "4"),
			gomc.NewRequest(4, "Propose", "5"),
			gomc.NewRequest(5, "Propose", "6"),
		),
		gomc.WithStateManager[GrpcConsensus, state](sm),
		gomc.WithPredicate(predicates...),
		gomc.IncorrectNodes(func(t *GrpcConsensus) { t.Stop() }, 3, 5),
	)
	sm.Export(os.Stdout)

	fmt.Println()
	ok, text := resp.Response()
	if !ok {
		t.Errorf("Expected simulation to pass. Got:\n %v", text)

		var buffer bytes.Buffer
		json.NewEncoder(&buffer).Encode(resp.Export())
		os.WriteFile("FailedRun.txt", buffer.Bytes(), 0755)
	}
}

func TestReplayConsensus(t *testing.T) {
	in, err := os.ReadFile("FailedRun.txt")
	if err != nil {
		t.Errorf("Error while setting up test: %v", err)
	}
	buffer := bytes.NewBuffer(in)
	var run []string
	json.NewDecoder(buffer).Decode(&run)
	sim := gomc.Prepare[GrpcConsensus, state](gomc.ReplayScheduler(run))

	sm := gomc.NewTreeStateManager(
		func(node *GrpcConsensus) state {
			decided := make([]string, len(node.DecidedVal))
			copy(decided, node.DecidedVal)
			return state{
				proposed: node.ProposedVal,
				decided:  decided,
			}
		},
		func(a, b state) bool {
			if a.proposed != b.proposed {
				return false
			}
			return slices.Equal(a.decided, b.decided)
		},
	)

	ids := []int32{1, 2, 3, 4, 5}
	addrMap := map[int32]string{
		1: ":50000",
		2: ":50001",
		3: ":50002",
		4: ":50003",
		5: ":50004",
	}

	addrToIdMap := map[string]int{}
	for id, addr := range addrMap {
		addrToIdMap[addr] = int(id)
	}

	resp := sim.RunSimulation(
		gomc.InitNodeFunc(
			func(sp gomc.SimulationParameters) map[int]*GrpcConsensus {

				gem := gomcGrpc.NewGrpcEventManager(addrToIdMap, sp.Sch, sp.NextEvt)

				lisMap := map[string]*bufconn.Listener{}
				for _, addr := range addrMap {
					lisMap[addr] = bufconn.Listen(bufSize)
				}

				nodes := map[int]*GrpcConsensus{}
				for id, addr := range addrMap {
					gc := NewGrpcConsensus(id, lisMap[addr], gem.WaitForSend)
					sp.Fm.Subscribe(gc.Crash)
					nodes[int(id)] = gc
				}

				for id, node := range nodes {
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
						grpc.WithUnaryInterceptor(gem.UnaryClientControllerInterceptor(int(id))),
					)
				}
				return nodes
			},
		),
		gomc.WithRequests(
			gomc.NewRequest(1, "Propose", "2"),
			gomc.NewRequest(2, "Propose", "3"),
			gomc.NewRequest(3, "Propose", "4"),
			gomc.NewRequest(4, "Propose", "5"),
			gomc.NewRequest(5, "Propose", "6"),
		),
		gomc.WithStateManager[GrpcConsensus, state](sm),
		gomc.WithPredicate(predicates...),
		gomc.IncorrectNodes(func(t *GrpcConsensus) { t.Stop() }, 3, 5),
	)
	sm.Export(os.Stdout)

	fmt.Println()
	ok, text := resp.Response()
	if !ok {
		t.Errorf("Expected simulation to pass. Got:\n %v", text)
	}
}
