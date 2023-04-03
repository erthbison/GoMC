package controller

import (
	"context"
	"fmt"
	"gomc/runner/recorder"
	"strconv"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type NodeController interface {
	// Pause the execution of a node
	Pause(int) error
	// Resume the execution of a node
	Resume(int) error
}

type GrpcNodeController struct {
	wait   map[int]*sync.WaitGroup
	paused map[int]bool

	msgChan chan recorder.Message

	lockMap   map[int]*sync.Mutex
	AddrIdMap map[string]int
}

func NewGrpcNodeController(addr map[string]int) *GrpcNodeController {
	wait := make(map[int]*sync.WaitGroup)
	msgChan := make(chan recorder.Message)

	lockMap := make(map[int]*sync.Mutex)
	for _, id := range addr {
		wait[id] = new(sync.WaitGroup)
		lockMap[id] = new(sync.Mutex)
	}
	return &GrpcNodeController{
		wait:      wait,
		paused:    make(map[int]bool),
		AddrIdMap: addr,
		msgChan:   msgChan,
		lockMap:   lockMap,
	}
}

func (gnc *GrpcNodeController) Stop() {
	close(gnc.msgChan)
}

func (gnc *GrpcNodeController) Subscribe() <-chan recorder.Message {
	return gnc.msgChan
}

func (gnc *GrpcNodeController) Pause(id int) error {
	if gnc.paused[id] {
		return fmt.Errorf("Controller: Node already paused")
	}
	gnc.paused[id] = true
	gnc.wait[id].Add(1)
	return nil
}

func (gnc *GrpcNodeController) Resume(id int) error {
	if !gnc.paused[id] {
		return fmt.Errorf("Controller: Node is not paused")
	}
	gnc.paused[id] = false
	gnc.wait[id].Done()
	return nil
}

func (gnc *GrpcNodeController) ServerInterceptor(id int) grpc.UnaryServerInterceptor {
	wait := gnc.wait[id]
	// the mutex is shared between all interceptors on the same process
	// It ensures that the messages are executed in the order in which they are sent on the channel and that they can not be interrupted during execution
	l := gnc.lockMap[id]
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		from, err := getId(ctx)
		if err != nil {
			panic(err)
		}
		wait.Wait()

		l.Lock()
		defer l.Unlock()
		gnc.msgChan <- recorder.Message{
			From: from,
			To:   id,
			Sent: false,
			Msg:  req,
		}
		return handler(ctx, req)
	}
}

func (gnc *GrpcNodeController) ClientInterceptor(id int) grpc.UnaryClientInterceptor {

	// the mutex is shared between all interceptors on the same process
	// It ensures that the messages are executed in the order in which they are sent on the channel and that they can not be interrupted during execution
	l := gnc.lockMap[id]
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.AppendToOutgoingContext(ctx, "gomc-id", strconv.Itoa(id))
		target, ok := gnc.AddrIdMap[cc.Target()] // HACK: cc.Target() is an experimental API
		if !ok {
			panic("Address of target not in provided address to id map")
		}

		l.Lock()
		gnc.msgChan <- recorder.Message{
			From: id,
			To:   target,
			Sent: true,
			Msg:  req,
		}
		l.Unlock()
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func getId(ctx context.Context) (int, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, fmt.Errorf("No metadata to retrieve")
	}
	vals := md.Get("gomc-id")
	if len(vals) > 1 {
		return 0, fmt.Errorf("To many id's in metadata")
	}
	if len(vals) == 0 {
		return 0, fmt.Errorf("No id's in metadata")
	}
	return strconv.Atoi(vals[0])
}
