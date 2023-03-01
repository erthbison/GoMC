package runner

import (
	"net"
	"reflect"
	"time"

	"google.golang.org/grpc/test/bufconn"
)

type Node interface {
	Start(net.Listener, func(string) (net.Conn, error))
	State() any
}

type Runner struct {
	ticker *time.Ticker
	nodes  map[int]Node

	stateChannels []chan map[int]any
}

func NewRunner(pollingInterval time.Duration) *Runner {
	return &Runner{
		ticker:        time.NewTicker(pollingInterval),
		stateChannels: make([]chan map[int]any, 0),
	}
}

func (r *Runner) Start(createNode func(id int) Node, addrs map[int]string) {
	nodes := map[int]Node{}
	listeners := map[string]*bufconn.Listener{}
	for id, addr := range addrs {
		nodes[id] = createNode(id)
		lis := bufconn.Listen(1024)
		listeners[addr] = lis
	}
	r.nodes = nodes

	dial := func(s string) (net.Conn, error) {
		return listeners[s].Dial()
	}

	for id, node := range nodes {
		go node.Start(listeners[addrs[id]], dial)
	}

	for range r.ticker.C {
		states := make(map[int]any)
		for id, node := range nodes {
			states[id] = node.State()
		}
		for _, chn := range r.stateChannels {
			chn <- states
		}
	}
}

func (r *Runner) Stop() {
	r.ticker.Stop()
}

// Subscribe to state updates
func (r *Runner) GetStateUpdates() chan map[int]any {
	chn := make(chan map[int]any)
	r.stateChannels = append(r.stateChannels, chn)
	return chn
}

func (r *Runner) Request(id int, requestType string, params ...any) {
	node := r.nodes[id]
	valueParams := make([]reflect.Value, len(params))
	for i, val := range params {
		valueParams[i] = reflect.ValueOf(val)
	}
	method := reflect.ValueOf(node).MethodByName(requestType)
	method.Call(valueParams)
}
