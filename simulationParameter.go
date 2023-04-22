package gomc

import (
	"gomc/eventManager"
)

type SimulationParameters struct {
	NextEvt        func(error, int)
	CrashSubscribe func(func(id int, status bool))
	EventAdder     eventManager.EventAdder
}
