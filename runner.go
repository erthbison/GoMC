package gomc

import (
	"fmt"
	"gomc/event"
	"gomc/runner"
	"reflect"
	"sync"
	"time"
)

type Runner[T, S any] struct {
	sync.Mutex

	interval      time.Duration
	stateChannels []chan map[int]S

	stopFunc func(*T) error

	requestSeq int
	ec         *runner.EventController[T, S]

	stateSubscribe chan chan map[int]S
	cmd            chan interface{}
	resp           chan error

	crashSubscribes []func(id int, status bool)
}

func NewRunner[T, S any](pollingInterval time.Duration, stop func(*T) error) *Runner[T, S] {
	return &Runner[T, S]{
		interval:      pollingInterval,
		stateChannels: make([]chan map[int]S, 0),
		stopFunc:      stop,

		ec: runner.NewEventController[T, S](),

		stateSubscribe: make(chan chan map[int]S),
		cmd:            make(chan interface{}),
		resp:           make(chan error),

		crashSubscribes: make([]func(id int, status bool), 0),
	}
}

func (r *Runner[T, S]) CrashSubscribe(f func(id int, status bool)) {
	r.crashSubscribes = append(r.crashSubscribes, f)
}

func (r *Runner[T, S]) Start(initNodes func(sp SimulationParameters) map[int]*T, getState func(*T) S) {

	nodes := initNodes(SimulationParameters{
		CrashSubscribe: r.CrashSubscribe,
		EventAdder:     r.ec,
		NextEvt:        r.ec.NextEvent,
	})

	r.ec.MainLoop(nodes, r.stopFunc, getState)

	go func() {
		ticker := time.NewTicker(r.interval)
		for cmd := range r.cmd {
			var err error
			switch t := cmd.(type) {
			case runner.Pause:
				err = r.pauseNode(t.Id, nodes)
			case runner.Resume:
				err = r.resumeNode(t.Id, nodes)
			case runner.Crash:
				err = r.crashNode(t.Id, nodes)
			case runner.Request:
				err = r.request(t.Id, t.Method, t.Params, nodes)
			case runner.Stop:
				err = r.stop(ticker, nodes)
			}
			r.resp <- err
		}
	}()
}

func (r *Runner[T, S]) SubscribeMessages() <-chan runner.Record {
	return r.ec.Subscribe()
}

func (r *Runner[T, S]) Stop() error {
	r.cmd <- runner.Stop{}
	return <-r.resp
}

func (r *Runner[T, S]) stop(ticker *time.Ticker, nodes map[int]*T) error {
	ticker.Stop()
	for _, n := range nodes {
		r.stopFunc(n)
	}
	for _, c := range r.stateChannels {
		close(c)
	}
	r.ec.Stop()
	close(r.cmd)
	return nil
}

func (r *Runner[T, S]) Request(id int, requestType string, params ...any) error {
	reflectParams := make([]reflect.Value, len(params))
	for i, val := range params {
		reflectParams[i] = reflect.ValueOf(val)
	}
	r.cmd <- runner.Request{
		Id:     id,
		Method: requestType,
		Params: reflectParams,
	}
	return <-r.resp
}

func (r *Runner[T, S]) request(id int, method string, params []reflect.Value, nodes map[int]*T) error {
	_, ok := nodes[id]
	if !ok {
		return fmt.Errorf("Invalid Node id")
	}
	r.ec.AddEvent(event.NewFunctionEvent(r.requestSeq, id, method, params...))
	r.requestSeq++
	return nil
}

func (r *Runner[T, S]) PauseNode(id int) error {
	r.cmd <- runner.Pause{
		Id: id,
	}
	return <-r.resp
}

func (r *Runner[T, S]) pauseNode(id int, nodes map[int]*T) error {
	_, ok := nodes[id]
	if !ok {
		return fmt.Errorf("Invalid Node id")
	}
	return r.ec.Pause(id)
}

func (r *Runner[T, S]) ResumeNode(id int) error {
	r.cmd <- runner.Resume{
		Id: id,
	}
	return <-r.resp
}

func (r *Runner[T, S]) resumeNode(id int, nodes map[int]*T) error {
	_, ok := nodes[id]
	if !ok {
		return fmt.Errorf("Invalid node Id")
	}
	return r.ec.Resume(id)
}

func (r *Runner[T, S]) CrashNode(id int) error {
	r.cmd <- runner.Crash{
		Id: id,
	}
	return <-r.resp
}

func (r *Runner[T, S]) crashNode(id int, nodes map[int]*T) error {
	n, ok := nodes[id]
	if !ok {
		return fmt.Errorf("Invalid node id")
	}
	for _, f := range r.crashSubscribes {
		f(id, false)
	}

	r.ec.CrashNode(id)
	return r.stopFunc(n)
}
