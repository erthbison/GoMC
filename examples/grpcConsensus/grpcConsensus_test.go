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
	"gomc/checking"
	"gomc/event"
	"gomc/gomcGrpc"

	"golang.org/x/exp/slices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

var predicates = []checking.Predicate[state]{
	checking.Eventually(
		// C1: Termination
		func(s checking.State[state]) bool {
			return checking.ForAllNodes(func(s state) bool {
				return len(s.decided) > 0
			}, s, true)
		},
	),
	func(s checking.State[state]) bool {
		// C2: Validity
		proposed := make(map[string]bool)
		for _, node := range s.LocalStates {
			proposed[node.proposed] = true
		}
		return checking.ForAllNodes(func(s state) bool {
			if len(s.decided) < 1 {
				// The process has not decided a value yet
				return true
			}
			return proposed[s.decided[0]]
		}, s, false)
	},
	func(s checking.State[state]) bool {
		// C3: Integrity
		return checking.ForAllNodes(func(s state) bool { return len(s.decided) < 2 }, s, false)
	},
	func(s checking.State[state]) bool {
		// C4: Agreement
		decided := make(map[string]bool)
		checking.ForAllNodes(func(s state) bool {
			for _, val := range s.decided {
				decided[val] = true
			}
			return true
		}, s, true)
		return len(decided) <= 1
	},
}

var addrMap = map[int32]string{
	1: ":50000",
	2: ":50001",
	3: ":50002",
	4: ":50003",
}

func TestGrpcConsensus(t *testing.T) {
	sim := gomc.PrepareSimulation(
		gomc.WithTreeStateManager(
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
		),
		gomc.RandomWalkScheduler(1),
	)

	addrToIdMap := map[string]int{}
	for id, addr := range addrMap {
		addrToIdMap[addr] = int(id)
	}

	resp := sim.Run(
		gomc.InitNodeFunc(
			func(sp gomc.SimulationParameters) map[int]*GrpcConsensus {
				gem := gomcGrpc.NewGrpcEventManager(addrToIdMap, sp.EventAdder, sp.NextEvt)
				lisMap := map[string]*bufconn.Listener{}
				for _, addr := range addrMap {
					lisMap[addr] = bufconn.Listen(bufSize)
				}

				nodes := map[int]*GrpcConsensus{}
				for id, addr := range addrMap {
					gc := NewGrpcConsensus(id, lisMap[addr], gem.WaitForSend(int(id)))
					sp.CrashSubscribe(int(id), gc.Crash)
					nodes[int(id)] = gc
				}

				for id, node := range nodes {
					node.DialServers(
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
			gomc.NewRequest(1, "Propose", "1"),
			gomc.NewRequest(2, "Propose", "2"),
			gomc.NewRequest(3, "Propose", "3"),
			gomc.NewRequest(4, "Propose", "4"),
		),
		gomc.WithPredicateChecker(predicates...),
		gomc.WithPerfectFailureManager(
			func(t *GrpcConsensus) { t.Stop() }, 2,
		),
		gomc.WithStopFunction(func(t *GrpcConsensus) { t.Stop() }),
	)

	fmt.Println()
	ok, text := resp.Response()
	if ok {
		t.Errorf("Expected simulation to fail. Got:\n %v", text)
	} else {
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
	var run []event.EventId
	json.NewDecoder(buffer).Decode(&run)

	sim := gomc.PrepareSimulation(
		gomc.WithTreeStateManager(
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
		),
		gomc.ReplayScheduler(run),
	)

	addrToIdMap := map[string]int{}
	for id, addr := range addrMap {
		addrToIdMap[addr] = int(id)
	}

	resp := sim.Run(
		gomc.InitNodeFunc(
			func(sp gomc.SimulationParameters) map[int]*GrpcConsensus {
				gem := gomcGrpc.NewGrpcEventManager(addrToIdMap, sp.EventAdder, sp.NextEvt)
				lisMap := map[string]*bufconn.Listener{}
				for _, addr := range addrMap {
					lisMap[addr] = bufconn.Listen(bufSize)
				}

				nodes := map[int]*GrpcConsensus{}
				for id, addr := range addrMap {
					gc := NewGrpcConsensus(id, lisMap[addr], gem.WaitForSend(int(id)))
					sp.CrashSubscribe(int(id), gc.Crash)
					nodes[int(id)] = gc
				}

				for id, node := range nodes {
					node.DialServers(
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
			gomc.NewRequest(1, "Propose", "1"),
			gomc.NewRequest(2, "Propose", "2"),
			gomc.NewRequest(3, "Propose", "3"),
			gomc.NewRequest(4, "Propose", "4"),
		),
		gomc.WithPredicateChecker(predicates...),
		gomc.WithPerfectFailureManager(
			func(t *GrpcConsensus) { t.Stop() }, 2,
		),
		gomc.WithStopFunction(func(t *GrpcConsensus) { t.Stop() }),
	)

	fmt.Println()
	ok, text := resp.Response()
	if ok {
		t.Errorf("Expected simulation to fail. Got:\n %v", text)
	}
}

func BenchmarkConsensus(b *testing.B) {
	addrMap := map[int32]string{
		1: ":50000",
		2: ":50001",
		3: ":50002",
		4: ":50003",
	}

	addrToIdMap := map[string]int{}
	for id, addr := range addrMap {
		addrToIdMap[addr] = int(id)
	}

	for i := 0; i < b.N; i++ {
		sim := gomc.PrepareSimulation(
			gomc.WithTreeStateManager(
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
			),
			gomc.PrefixScheduler(),
		)

		sim.Run(
			gomc.InitNodeFunc(
				func(sp gomc.SimulationParameters) map[int]*GrpcConsensus {
					gem := gomcGrpc.NewGrpcEventManager(addrToIdMap, sp.EventAdder, sp.NextEvt)
					lisMap := map[string]*bufconn.Listener{}
					for _, addr := range addrMap {
						lisMap[addr] = bufconn.Listen(bufSize)
					}

					nodes := map[int]*GrpcConsensus{}
					for id, addr := range addrMap {
						gc := NewGrpcConsensus(id, lisMap[addr], gem.WaitForSend(int(id)))
						sp.CrashSubscribe(int(id), gc.Crash)
						nodes[int(id)] = gc
					}

					for id, node := range nodes {
						node.DialServers(
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
				gomc.NewRequest(1, "Propose", "1"),
				gomc.NewRequest(2, "Propose", "2"),
				gomc.NewRequest(3, "Propose", "3"),
				gomc.NewRequest(4, "Propose", "4"),
			),
			gomc.WithPredicateChecker(predicates...),
			gomc.WithPerfectFailureManager(
				func(t *GrpcConsensus) { t.Stop() }, 2,
			),
			gomc.WithStopFunction(func(t *GrpcConsensus) { t.Stop() }),
		)
	}
}
