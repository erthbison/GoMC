package scheduler

import "testing"

func TestQueueSchedulerCrash(t *testing.T) {
	sch := NewQueueScheduler()
	testSchedulerCrash(sch, t)
}

func TestQueueSchedulerExplore2Events(t *testing.T) {
	sch := NewQueueScheduler()
	testDeterministicExplore2Events(t, sch)
}

func TestQueueExploreBranchingEvents(t *testing.T) {
	sch := NewQueueScheduler()
	testDeterministicExploreBranchingEvents(t, sch)
}

func BenchmarkQueueScheduler(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sch := NewQueueScheduler()
		err := benchmarkScheduler(sch, 7)
		if err != nil {
			b.Errorf(err.Error())
		}
	}
}
