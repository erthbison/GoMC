package gomc

import "gomc/scheduler"

type SimulationParameters struct {
	NextEvt   chan error
	Subscribe func(func(id int, status bool))
	Sch       scheduler.RunScheduler
}
