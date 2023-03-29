package runnerControllers

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/grpc"
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
}

func NewGrpcNodeController(nodeIds []int) *GrpcNodeController {
	wait := make(map[int]*sync.WaitGroup)
	for _, id := range nodeIds {
		wait[id] = new(sync.WaitGroup)
	}
	return &GrpcNodeController{
		wait:   wait,
		paused: make(map[int]bool),
	}
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
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		gnc.wait[id].Wait()
		return handler(ctx, req)
	}
}
