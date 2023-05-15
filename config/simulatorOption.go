package config

import "gomc/scheduler"

type SchedulerOption struct {
	Sch scheduler.GlobalScheduler
}

func (so SchedulerOption) SimOpt() {}

type MaxDepthOption struct{ MaxDepth int }

func (mdo MaxDepthOption) SimOpt() {}

type MaxRunsOption struct{ MaxRuns int }

func (mro MaxRunsOption) SimOpt() {}

type NumConcurrentOption struct{ N int }

func (nco NumConcurrentOption) SimOpt() {}

type IgnorePanicOption struct{}

func (ipo IgnorePanicOption) SimOpt() {}

type IgnoreErrorOption struct{}

func (ieo IgnoreErrorOption) SimOpt() {}
