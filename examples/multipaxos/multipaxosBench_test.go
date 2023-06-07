package multipaxos

import (
	"gomc"
	"testing"
)

var nodes = map[int64]string{
	1: ":1",
	2: ":2",
	3: ":3",
}

func BenchmarkMultipaxos(b *testing.B) {
	sim := gomc.PrepareSimulation(
		gomc.WithTreeStateManager(getState, cmpState),
		gomc.PrefixScheduler(),
	)

	for i := 0; i < b.N; i++ {
		sim.Run(
			gomc.InitNodeFunc(
				InitNodes(nodes),
			),
			gomc.WithRequests(
				gomc.NewRequest(1, "ProposeVal", "1"),
				gomc.NewRequest(2, "ProposeVal", "2"),
				gomc.NewRequest(3, "ProposeVal", "3"),
			),
			gomc.WithPredicateChecker(predicates...),
			gomc.WithPerfectFailureManager(
				func(t *MultiPaxos) { t.Stop() }, 2,
			),
			gomc.WithStopFunctionSimulator(func(t *MultiPaxos) { t.Stop() }),
		)
	}
}
