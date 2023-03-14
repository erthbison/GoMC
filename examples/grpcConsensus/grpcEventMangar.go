package main

import (
	"context"
	"fmt"
	"gomc/scheduler"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type grpcEventManager struct {
	AddrIdMap map[string]int
	sch       scheduler.Scheduler
	nextEvt   chan error

	wait map[int]chan bool
}

func NewGrpcEventManager(addr map[string]int, sch scheduler.Scheduler, nextEvent chan error) *grpcEventManager {
	wait := make(map[int]chan bool)
	for _, id := range addr {
		wait[id] = make(chan bool)
	}
	return &grpcEventManager{
		AddrIdMap: addr,
		sch:       sch,
		nextEvt:   nextEvent,
		wait:      wait,
	}
}

func (gem *grpcEventManager) addEvent(from, to int, msg interface{}, method string)  {
	gem.sch.AddEvent(GrpcEvent{
		target: to,
		from:   from,
		method: method,
		msg:    msg,
		wait:   gem.wait[to],
	})
}

// func (gem *grpcEventManager) NodeCrash(id int) {
// 	close(gem.wait[id])
// }

// Called when the client receives a response from the server
func (gem *grpcEventManager) UnaryClientControllerInterceptor(id int) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.AppendToOutgoingContext(ctx, "gomc-node-id", strconv.Itoa(id))
		target := gem.AddrIdMap[cc.Target()] // HACK: cc.Target() is an experimental API

		// Create a request event and signal that the old event has been completed
		fmt.Println("Client", id, "To", target, "Req", req)
		gem.addEvent(id, target, req, method)

		// Wait for the created request event to be executed
		gem.nextEvt <- nil
		<-gem.wait[target]

		err := invoker(ctx, method, req, reply, cc, opts...)
		
		gem.addEvent(target, id, reply, method)
		gem.nextEvt <- nil

		// // Wait for the response event to execute
		fmt.Println("Client", id, "From", target, "Rep", reply)
		<-gem.wait[id]

		return err
	}
}

// Called when the server receives a request from the client
func (gem *grpcEventManager) UnaryServerControllerInterceptor(id int) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		from, err := gem.getSender(ctx)
		if err != nil {
			panic(err)
		}
		fmt.Println("Server", id, "From", from, "Req", req)

		// fmt.Println("Server", id, "Executing")
		// gem.nextEvt <- nil
		// <-gem.wait[id]
		resp, err = handler(ctx, req)
		// Create a response event and signal that the request event has been completed

		fmt.Println("Server", id, "To", from, "resp", resp)
		// err = gem.addEvent(id, from, resp, info.FullMethod)
		// if err != nil {
		// 	return nil, err
		// }

		// gem.nextEvt <- nil
		// <-gem.wait[from]

		return resp, err
	}
}

func (gem *grpcEventManager) getSender(ctx context.Context) (int, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, fmt.Errorf("Error while extracting metadata")
	}
	fromSlice := md.Get("gomc-node-id")
	if len(fromSlice) != 1 {
		return 0, fmt.Errorf("Error while extracting metadata")
	}
	from, err := strconv.Atoi(fromSlice[0])
	if err != nil {
		return 0, fmt.Errorf("Error while extracting metadata")
	}
	return from, nil
}
