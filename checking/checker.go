package checking

import "gomc/state"

type Checker[S any] interface {
	Check(state.StateSpace[S]) CheckerResponse
}

type CheckerResponse interface {
	Response() (bool, string)
	Export() []uint64
}
