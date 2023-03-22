package scheduler

import "testing"

func TestQueueSchedulerExplore2Events(t *testing.T) {
	sch := NewPrefix()
	testDeterministicExplore2Events(t, sch)
}

func TestQueueExploreBranchingEvents(t *testing.T) {
	sch := NewPrefix()
	testDeterministicExploreBranchingEvents(t, sch)
}

func TestConcurrentBranchingEvent(t *testing.T) {
	sch := NewPrefix()
	testConcurrentDeterministic(t, sch)
}

func BenchmarkQueueScheduler(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sch := NewPrefix()
		err := benchmarkRunScheduler(sch.GetRunScheduler(), 7)
		if err != nil {
			b.Errorf(err.Error())
		}
	}
}
