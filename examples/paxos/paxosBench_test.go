package paxos

import (
	"context"
	"gomc/eventManager"
	"net"
	"testing"

	"gomc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func BenchmarkPaxos(b *testing.B) {
	addresses := map[int64]string{
		1: ":1",
		2: ":2",
		3: ":3",
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

	addrToIdMap := map[string]int{}
	for id, addr := range addresses {
		addrToIdMap[addr] = int(id)
	}

	for i := 0; i < b.N; i++ {
		sim.Run(
			gomc.InitNodeFunc(func(sp eventManager.SimulationParameters) map[int]*Server {
				lisMap := map[string]*bufconn.Listener{}
				for _, addr := range addresses {
					lisMap[addr] = bufconn.Listen(bufSize)
				}
				gem := eventManager.NewGrpcEventManager(addrToIdMap, sp)

				nodes := make(map[int]*Server)
				for id, addr := range addresses {
					srv, err := NewServer(id, addresses, gem.WaitForSend(int(id)))
					if err != nil {
						b.Errorf("Error while starting simulation: %v", err)
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
			),
			gomc.WithPredicateChecker(predicates...),
			gomc.WithPerfectFailureManager(func(t *Server) { t.Stop() }, 1),
			gomc.WithStopFunctionSimulator(func(t *Server) { t.Stop() }),
		)
	}
}
