package gomc

import (
	"fmt"
	"gomc/runner"
	"net"
	"reflect"
	"time"

	"google.golang.org/grpc/test/bufconn"
)

type Dialer func(string) (net.Conn, error)

type Runner[T any] struct {
	ticker *time.Ticker
	nodes  map[int]*T

	stateChannels []chan map[int]any

	stop func(*T) error

	ctrl runner.NodeController
	fm   *failureManager
}

func NewRunner[T any](pollingInterval time.Duration, ctrl runner.NodeController, stop func(*T) error) *Runner[T] {
	return &Runner[T]{
		ticker:        time.NewTicker(pollingInterval),
		stateChannels: make([]chan map[int]any, 0),

		stop: stop,

		ctrl: ctrl,
		fm:   newFailureManager(),
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
	r.nodes = nodes
	r.fm.Init(nodeIds)

	dial := func(s string) (net.Conn, error) {
		return listeners[s].Dial()
	}

	for id, node := range nodes {
		go start(node, listeners[addrs[id]], dial)
	}

	for range r.ticker.C {
		states := make(map[int]any)
		for id, node := range nodes {
			states[id] = getState(node)
		}
		for _, chn := range r.stateChannels {
			chn <- states
		}
	}
}

func (r *Runner[T]) Stop() {
	r.ticker.Stop()
	for _, n := range r.nodes {
		r.stop(n)
	}
	for _, c := range r.stateChannels {
		close(c)
	}
}

// Subscribe to state updates
func (r *Runner[T]) GetStateUpdates() chan map[int]any {
	chn := make(chan map[int]any)
	r.stateChannels = append(r.stateChannels, chn)
	return chn
}

func (r *Runner[T]) Request(id int, requestType string, params ...any) {
	node := r.nodes[id]
	valueParams := make([]reflect.Value, len(params))
	for i, val := range params {
		valueParams[i] = reflect.ValueOf(val)
	}
	method := reflect.ValueOf(node).MethodByName(requestType)
	method.Call(valueParams)
}

func (r *Runner[T]) PauseNode(id int) error {
	_, ok := r.nodes[id]
	if !ok {
		return fmt.Errorf("Invalid Node id")
	}
	return r.ctrl.Pause(id)
}

func (r *Runner[T]) ResumeNode(id int) error {
	_, ok := r.nodes[id]
	if !ok {
		return fmt.Errorf("Invalid node Id")
	}
	return r.ctrl.Resume(id)
}

func (r *Runner[T]) CrashNode(id int) error {
	n, ok := r.nodes[id]
	if !ok {
		return fmt.Errorf("Invalid node id")
	}
	err := r.fm.NodeCrash(id)
	if err != nil {
		return err
	}
	return r.stop(n)
}

func (r *Runner[T]) CrashSubscribe(callback func(int)) {
	r.fm.Subscribe(callback)
}
