package gomc

import (
	"gomc/runner"
	"sync"
)

type Runner[T, S any] struct {
	sync.Mutex

	rc *runner.RunnerController[T, S]

	cmd            chan interface{}
	resp           chan error
}

func NewRunner[T, S any](recordChanBuffer int) *Runner[T, S] {
	return &Runner[T, S]{
		rc: runner.NewEventController[T, S](recordChanBuffer),

		cmd:            make(chan interface{}),
		resp:           make(chan error),
	}
}

func (r *Runner[T, S]) Start(initNodes func(sp SimulationParameters) map[int]*T, getState func(*T) S, stop func(*T), eventChanBuffer int) {

	nodes := initNodes(SimulationParameters{
		CrashSubscribe: r.rc.CrashSubscribe,
		EventAdder:     r.rc,
		NextEvt:        r.rc.NextEvent,
	})

	r.rc.MainLoop(nodes, eventChanBuffer, stop, getState)

	go func() {
		for cmd := range r.cmd {
			var err error
			switch t := cmd.(type) {
			case runner.Pause:
				err = r.rc.Pause(t.Id)
			case runner.Resume:
				err = r.rc.Resume(t.Id)
			case runner.Crash:
				err = r.rc.CrashNode(t.Id)
			case runner.Request:
				err = r.rc.NewRequest(t.Id, t.Method, t.Params)
			case runner.Stop:
				err = r.rc.Stop()
				close(r.cmd)
			}
			r.resp <- err
		}
	}()
}

func (r *Runner[T, S]) SubscribeMessages() <-chan runner.Record {
	return r.rc.Subscribe()
}

func (r *Runner[T, S]) Stop() error {
	r.cmd <- runner.Stop{}
	return <-r.resp
}

func (r *Runner[T, S]) Request(req Request) error {
	r.cmd <- runner.Request{
		Id:     req.Id,
		Method: req.Method,
		Params: req.Params,
	}
	return <-r.resp
}

func (r *Runner[T, S]) PauseNode(id int) error {
	r.cmd <- runner.Pause{
		Id: id,
	}
	return <-r.resp
}

func (r *Runner[T, S]) ResumeNode(id int) error {
	r.cmd <- runner.Resume{
		Id: id,
	}
	return <-r.resp
}

func (r *Runner[T, S]) CrashNode(id int) error {
	r.cmd <- runner.Crash{
		Id: id,
	}
	return <-r.resp
}
