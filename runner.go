package gomc

import (
	"fmt"
	"gomc/failureManager"
	"gomc/runner"
	"gomc/runner/controller"
	"gomc/runner/recorder"
	"reflect"
	"sync"
	"time"
)

type Runner[T, S any] struct {
	sync.Mutex

	interval time.Duration

	stateChannels []chan map[int]S

	stopFunc func(*T) error

	rec  recorder.MessageRecorder
	ctrl controller.NodeController
	fm   *failureManager.FailureManager

	stateSubscribe chan chan map[int]S
	cmd            chan interface{}
	resp           chan error
}

func NewRunner[T, S any](pollingInterval time.Duration, ctrl controller.NodeController, rec recorder.MessageRecorder, stop func(*T) error) *Runner[T, S] {
	return &Runner[T, S]{
		interval:      pollingInterval,
		stateChannels: make([]chan map[int]S, 0),
		stopFunc:      stop,

		rec:  rec,
		ctrl: ctrl,
		fm:   failureManager.New(),

		stateSubscribe: make(chan chan map[int]S),
		cmd:            make(chan interface{}),
		resp:           make(chan error),
	}
}

func (r *Runner[T, S]) Start(initNodes func(sp SimulationParameters) map[int]*T, getState func(*T) S) {
	nodes := initNodes(SimulationParameters{
		Fm: r.fm,
	})
	nodeIds := []int{}
	for id := range nodes {
		nodeIds = append(nodeIds, id)
	}
	r.fm.Init(nodeIds)

	go func() {
		ticker := time.NewTicker(r.interval)
		run := true
		for run {
			select {
			case <-ticker.C:
				states := make(map[int]S)
				for id, node := range nodes {
					states[id] = getState(node)
				}
				for _, chn := range r.stateChannels {
					chn <- states
				}
			case cmd := <-r.cmd:
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
					if err == nil {
						run = false
					}
				}
				r.resp <- err
			case c := <-r.stateSubscribe:
				r.stateChannels = append(r.stateChannels, c)
			}
		}
	}()
}

func (r *Runner[T, S]) SubscribeMessages() <-chan recorder.Message {
	return r.rec.Subscribe()
}

// Subscribe to state updates
func (r *Runner[T, S]) SubscribeStateUpdates() chan map[int]S {
	chn := make(chan map[int]S)
	r.stateSubscribe <- chn
	return chn
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
	node, ok := nodes[id]
	if !ok {
		return fmt.Errorf("Invalid Node id")
	}
	meth := reflect.ValueOf(node).MethodByName(method)
	meth.Call(params)
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
	return r.ctrl.Pause(id)
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
	return r.ctrl.Resume(id)
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
	err := r.fm.NodeCrash(id)
	if err != nil {
		return err
	}
	return r.stopFunc(n)
}
