package scheduler

import (
	"errors"
	"testing"
)

func TestBasicSchedulerNoEvents(t *testing.T) {
	sch := NewBasicScheduler()
	_, err := sch.GetEvent()
	if !errors.Is(err, RunEndedError) {
		t.Fatalf("unexpected error. Got %v. Expected: %v", err, RunEndedError)
	}
	_, err = sch.GetEvent()
	if !errors.Is(err, RunEndedError) {
		t.Fatalf("unexpected error. Got %v. Expected: %v", err, RunEndedError)
	}
	sch.EndRun()
}

func TestBasicSchedulerExplore2Events(t *testing.T) {
	sch := NewBasicScheduler()
	testDeterministicExplore2Events(t, sch)
}

func TestBasicExploreBranchingEvents(t *testing.T) {
	sch := NewBasicScheduler()
	testDeterministicExploreBranchingEvents(t, sch)
}

func TestBasicSchedulerNodeCrash(t *testing.T) {
	sch := NewBasicScheduler()
	testSchedulerCrash(sch, t)
}

func BenchmarkBasicScheduler(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sch := NewBasicScheduler()
		err := benchmarkScheduler(sch, 7)
		if err != nil {
			b.Errorf(err.Error())
		}
	}
}
