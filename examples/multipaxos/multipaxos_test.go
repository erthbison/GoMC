package multipaxos

import (
	"context"
	"net"
	"os"
	"testing"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"gomc"
	"gomc/checking"
	"gomc/eventManager"
)

var predicates = []checking.Predicate[State]{
	func(s checking.State[State]) bool {
		// Only a value that has been proposed may be chosen
		proposed := make(map[string]bool)
		for _, s := range s.LocalStates {
			for _, val := range s.proposed {
				proposed[val] = true
			}
		}
		return checking.ForAllNodes(func(s State) bool {
			for _, val := range s.decided {
				if !proposed[val] {
					return false
				}
			}
			return true
		}, s, false)
	},
	checking.Eventually(
		func(s checking.State[State]) bool {
			// All correct nodes decide on the same values for the same slots
			first := true
			var decidedVal map[int64]string
			for id, state := range s.LocalStates {
				if !s.Correct[id] {
					continue
				}
				if first {
					decidedVal = state.decided
					first = false
					continue
				}
				if !maps.Equal(decidedVal, state.decided) {
					return false
				}
			}

			return true
		},
	),
	// No liveness properties.
}

type State struct {
	proposed []string
	decided  map[int64]string
}

var (
	bufSize = 1024
	addrMap = map[int64]string{
		1: ":50000",
		2: ":50001",
		3: ":50002",
	}
	addrToIdMap = map[string]int{}
)

func getState(t *MultiPaxos) State {
	return State{
		proposed: slices.Clone(t.proposed),
		decided:  maps.Clone(t.LearntValues),
	}
}

func cmpState(s1, s2 State) bool {
	if !maps.Equal(s1.decided, s2.decided) {
		return false
	}
	return slices.Equal(s1.proposed, s2.proposed)
}

func InitNodes(addrMap map[int64]string) func(sp eventManager.SimulationParameters) map[int]*MultiPaxos {
	return func(sp eventManager.SimulationParameters) map[int]*MultiPaxos {
		lisMap := map[string]*bufconn.Listener{}
		for _, addr := range addrMap {
			lisMap[addr] = bufconn.Listen(bufSize)
		}
		gem := eventManager.NewGrpcEventManager(addrToIdMap, sp)

		nodes := make(map[int]*MultiPaxos)
		for id, addr := range addrMap {
			srv := NewMultiPaxos(id, addrMap, gem.WaitForSend(int(id)))
			go srv.Start(lisMap[addr])
			sp.CrashSubscribe(int(id), srv.proposer.leader.NodeCrash)
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
	}
}

func TestMultiPaxosSim(t *testing.T) {
	for id, addr := range addrMap {
		addrToIdMap[addr] = int(id)
	}

	sim := gomc.PrepareSimulation(
		gomc.WithTreeStateManager(getState, cmpState),
		gomc.RandomWalkScheduler(1),
		gomc.MaxRuns(10000),
	)
	w, err := os.Create("export.txt")
	if err != nil {
		t.Errorf("Error while creating file: %v", err)
	}
	defer w.Close()
	resp := sim.Run(
		gomc.InitNodeFunc(InitNodes(addrMap)),
		gomc.WithRequests(
			gomc.NewRequest(3, "ProposeVal", "Test1"),
			gomc.NewRequest(2, "ProposeVal", "Test2"),
			gomc.NewRequest(3, "ProposeVal", "Test3"),
			gomc.NewRequest(1, "ProposeVal", "Test4"),
		),
		gomc.WithPredicateChecker(predicates...),
		gomc.WithPerfectFailureManager(func(t *MultiPaxos) { t.Stop() }, 3),
		gomc.WithStopFunctionSimulator(func(t *MultiPaxos) { t.Stop() }),
		gomc.Export(w),
	)
	if ok, text := resp.Response(); !ok {
		t.Errorf("Test Failed: \n %v", text)
	}
}
