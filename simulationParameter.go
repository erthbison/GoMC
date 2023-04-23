package gomc

import (
	"gomc/eventManager"
)

type SimulationParameters struct {
	NextEvt        func(error, int)
	CrashSubscribe func(NodeId int, callback func(id int, status bool))
	EventAdder     eventManager.EventAdder
}
