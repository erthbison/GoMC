package runner

import (
	"gomc/eventManager"
	"gomc/request"
	"sync"
)

type Runner[T, S any] struct {
	sync.Mutex

	rc *RunnerController[T, S]

	cmd  chan command
	resp chan error
}

func NewRunner[T, S any](recordChanBuffer int) *Runner[T, S] {
	return &Runner[T, S]{
		rc: NewEventController[T, S](recordChanBuffer),

		cmd:  make(chan command),
		resp: make(chan error),
	}
}

func (r *Runner[T, S]) Start(initNodes func(sp eventManager.SimulationParameters) map[int]*T, getState func(*T) S, stop func(*T), eventChanBuffer int) {

	nodes := initNodes(eventManager.SimulationParameters{
		CrashSubscribe: r.rc.CrashSubscribe,
		EventAdder:     r.rc,
		NextEvt:        r.rc.NextEvent,
	})

	r.rc.MainLoop(nodes, eventChanBuffer, stop, getState)

	go func() {
		for cmd := range r.cmd {
			var err error
			switch t := cmd.(type) {
			case pauseCmd:
				err = r.rc.Pause(t.Id)
			case resumeCmd:
				err = r.rc.Resume(t.Id)
			case crashCmd:
				err = r.rc.CrashNode(t.Id)
			case requestCmd:
				err = r.rc.NewRequest(t.Id, t.Method, t.Params)
			case stopCmd:
				r.rc.Stop()
				close(r.cmd)
				err = nil
			}
			r.resp <- err
		}
	}()
}

// Subscribe to records of events and states that are reported by the nodes
//
// Events on different nodes can be executed concurrently, so order of events from different nodes does not necessarily match.
// Orders of events on the same node is guaranteed to match the order in which they where executed.
// Nodes send a MessageRecord when they either send or receive a message and an ExecutionRecord after they perform some local execution.
// Nodes send a StateRecord containing the new state of the node after they have received a message or executed some local event.
func (r *Runner[T, S]) SubscribeRecords() <-chan Record {
	return r.rc.Subscribe()
}

func (r *Runner[T, S]) Stop() error {
	r.cmd <- stopCmd{}
	return <-r.resp
}

func (r *Runner[T, S]) Request(req request.Request) error {
	r.cmd <- requestCmd{
		Id:     req.Id,
		Method: req.Method,
		Params: req.Params,
	}
	return <-r.resp
}

func (r *Runner[T, S]) PauseNode(id int) error {
	r.cmd <- pauseCmd{
		Id: id,
	}
	return <-r.resp
}

func (r *Runner[T, S]) ResumeNode(id int) error {
	r.cmd <- resumeCmd{
		Id: id,
	}
	return <-r.resp
}

func (r *Runner[T, S]) CrashNode(id int) error {
	r.cmd <- crashCmd{
		Id: id,
	}
	return <-r.resp
}
