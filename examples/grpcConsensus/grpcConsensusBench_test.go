package main

import (
	"gomc"
	"testing"
)

func BenchmarkConsensus(b *testing.B) {
	sim := gomc.PrepareSimulation(
		gomc.WithTreeStateManager(getState, cmpState),
		gomc.PrefixScheduler(),
	)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sim.Run(
			gomc.InitNodeFunc(
				createNodes(addrMap),
			),
			gomc.WithRequests(
				gomc.NewRequest(1, "Propose", "1"),
				gomc.NewRequest(2, "Propose", "2"),
				gomc.NewRequest(3, "Propose", "3"),
			),
			gomc.WithPredicateChecker(predicates...),
			gomc.WithPerfectFailureManager(
				func(t *GrpcConsensus) { t.Stop() }, 2,
			),
			gomc.WithStopFunctionSimulator(func(t *GrpcConsensus) { t.Stop() }),
		)
	}
}
