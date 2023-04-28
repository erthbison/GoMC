package checking

import (
	"gomc/event"
	"gomc/state"
)

type Checker[S any] interface {
	Check(root state.StateSpace[S]) CheckerResponse
}

type CheckerResponse interface {
	Response() (bool, string)
	Export() []event.EventId
}
