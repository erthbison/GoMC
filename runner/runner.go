package runner

import (
	"fmt"
	"gomc/failureManager"
	"gomc/runner/controller"
	"gomc/runner/recorder"
	"net"
	"reflect"
	"sync"
	"time"

	"google.golang.org/grpc/test/bufconn"
)

type Dialer func(string) (net.Conn, error)

type Runner[T any] struct {
	sync.Mutex

	interval time.Duration

	stateChannels []chan map[int]any

	stopFunc func(*T) error

	rec  recorder.MessageRecorder
	ctrl controller.NodeController
	fm   *failureManager.FailureManager

	stateSubscribe chan chan map[int]any
	cmd            chan interface{}
	resp           chan error
}

func NewRunner[T any](pollingInterval time.Duration, ctrl controller.NodeController, rec recorder.MessageRecorder, stop func(*T) error) *Runner[T] {
	return &Runner[T]{
		interval:      pollingInterval,
		stateChannels: make([]chan map[int]any, 0),
		stopFunc:      stop,

		rec:  rec,
		ctrl: ctrl,
		fm:   failureManager.New(),

		stateSubscribe: make(chan chan map[int]any),
		cmd:            make(chan interface{}),
		resp:           make(chan error),
	}
}

func (r *Runner[T]) Start(initNodes func() map[int]*T, addrs map[int]string, start func(*T, net.Listener, Dialer), getState func(*T) any) {
	nodes := initNodes()
	listeners := map[string]*bufconn.Listener{}
	nodeIds := []int{}
	for id, addr := range addrs {
		lis := bufconn.Listen(1024)
		listeners[addr] = lis
		nodeIds = append(nodeIds, id)
	}

	r.fm.Init(nodeIds)

	dial := func(s string) (net.Conn, error) {
		return listeners[s].Dial()
	}

	for id, node := range nodes {
		go start(node, listeners[addrs[id]], dial)
	}

	go func() {
		ticker := time.NewTicker(r.interval)
		run := true
		for run {
			select {
			case <-ticker.C:
				states := make(map[int]any)
				for id, node := range nodes {
					states[id] = getState(node)
				}
				for _, chn := range r.stateChannels {
					chn <- states
				}
			case cmd := <-r.cmd:
				var err error
				switch t := cmd.(type) {
				case pause:
					err = r.pauseNode(t.id, nodes)
				case resume:
					err = r.resumeNode(t.id, nodes)
				case crash:
					err = r.crashNode(t.id, nodes)
				case request:
					err = r.request(t.id, t.method, t.params, nodes)
				case stop:
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

func (r *Runner[T]) SubscribeMessages() <-chan recorder.Message {
	return r.rec.Subscribe()
}

// Subscribe to state updates
func (r *Runner[T]) SubscribeStateUpdates() chan map[int]any {
	chn := make(chan map[int]any)
	r.stateSubscribe <- chn
	return chn
}

func (r *Runner[T]) Stop() error {
	r.cmd <- stop{}
	return <-r.resp
}

func (r *Runner[T]) stop(ticker *time.Ticker, nodes map[int]*T) error {
	ticker.Stop()
	for _, n := range nodes {
		r.stopFunc(n)
	}
	for _, c := range r.stateChannels {
		close(c)
	}
	return nil
}

func (r *Runner[T]) Request(id int, requestType string, params ...any) error {
	reflectParams := make([]reflect.Value, len(params))
	for i, val := range params {
		reflectParams[i] = reflect.ValueOf(val)
	}
	r.cmd <- request{
		id:     id,
		method: requestType,
		params: reflectParams,
	}
	return <-r.resp
}

func (r *Runner[T]) request(id int, method string, params []reflect.Value, nodes map[int]*T) error {
	node, ok := nodes[id]
	if !ok {
		return fmt.Errorf("Invalid Node id")
	}
	meth := reflect.ValueOf(node).MethodByName(method)
	meth.Call(params)
	return nil
}

func (r *Runner[T]) PauseNode(id int) error {
	r.cmd <- pause{
		id: id,
	}
	return <-r.resp
}

func (r *Runner[T]) pauseNode(id int, nodes map[int]*T) error {
	_, ok := nodes[id]
	if !ok {
		return fmt.Errorf("Invalid Node id")
	}
	return r.ctrl.Pause(id)
}

func (r *Runner[T]) ResumeNode(id int) error {
	r.cmd <- resume{
		id: id,
	}
	return <-r.resp
}

func (r *Runner[T]) resumeNode(id int, nodes map[int]*T) error {
	_, ok := nodes[id]
	if !ok {
		return fmt.Errorf("Invalid node Id")
	}
	return r.ctrl.Resume(id)
}

func (r *Runner[T]) CrashNode(id int) error {
	r.cmd <- crash{
		id: id,
	}
	return <-r.resp
}

func (r *Runner[T]) crashNode(id int, nodes map[int]*T) error {
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

func (r *Runner[T]) CrashSubscribe(callback func(int)) {
	r.fm.Subscribe(callback)
}
