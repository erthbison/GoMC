package eventManager

import (
	"context"
	"errors"
	"gomc/event"
	"time"

	"google.golang.org/grpc"
)

// An Event Manager that will be used to control asynchronous messages sent using gRPC.
//
// Grpc async calls are called in a separate goroutine without handling the response.
// A context with a deadline should not be used when simulating since real time does not make sense during simulations.
//
// Uses a gRPC unary client interceptor to intercept and withhold messages.
// The interceptor is created by the UnaryClientControllerInterceptor method.
// 
// The function created by the WaitForSend method must be called after sending messages using gRPC.
// This ensures that all the messages are added to the EventAdder before continuing.
type GrpcEventManager struct {
	addrIdMap map[string]int

	ea      EventAdder
	nextEvt func(error, int)

	msgChan map[int]chan bool
}

// Create a GrpcEventManager for use when using grpc Async calls when simulating
// Grpc async calls are called in a separate goroutine without handling the response.
// A context with a deadline should not be used when simulating since real time does not make sense during simulations.
// addr is a map from address to node id, shc is the scheduler used and nextEvent is the NextEvent channel from the simulator
func NewGrpcEventManager(addr map[string]int, ea EventAdder, nextEvent func(error, int)) *GrpcEventManager {
	msgChan := make(map[int]chan bool)
	for _, id := range addr {
		msgChan[id] = make(chan bool)
	}
	return &GrpcEventManager{
		addrIdMap: addr,
		ea:        ea,
		nextEvt:   nextEvent,
		msgChan:   msgChan,
	}
}

// Add an grpcEvent to the scheduler.
func (gem *GrpcEventManager) addEvent(from, to int, msg interface{}, method string, wait chan bool) {
	gem.ea.AddEvent(event.NewGrpcEvent(
		from,
		to,
		method,
		msg,
		wait,
	))
}

// Creates a function that wait until all messages has been processed and an event has been created for all of them.
// id is the id of the node sending the messages
//
// The created function should be called right after performing an async multicast using grpc.
// num is the number of messages that are sent.
//
// The method ensures that all messages are added before the event is complete, ensuring that the simulation can proceed as expected and not finish prematurely or ignore some messages.
func (gem *GrpcEventManager) WaitForSend(id int) func(int) {
	return func(num int) {
		for i := 0; i < num; i++ {
			<-gem.msgChan[id]
		}
	}
}

// Create a UnaryClientInterceptor that is used to control the message flow of grpc events.
// The id is the id of the client node sending the requests
//
// It creates a GrpcEvent and holds the message until the grpcRequest event is executed.
// After the grpcRequest event is executed and an (empty) response has been received it signals that the event is completed and that the next event can be executed
func (gem *GrpcEventManager) UnaryClientControllerInterceptor(id int) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		target := gem.addrIdMap[cc.Target()] // HACK: cc.Target() is an experimental API

		// Create a request event
		wait := make(chan bool)
		gem.addEvent(id, target, req, method, wait)

		// Signal that an event has been created for the event

		select {
		case gem.msgChan[id] <- true:
		case <-time.After(10 * time.Second):
			// The wait for send method has not been called. Panic to i
			panic(errors.New("grpcEventManager: timed out while confirming that message has been processed. grpcEventManager.WaitForSend must be called after sending messages to ensure that the message is properly handled."))
		}
		// Wait until the event has been executed
		<-wait

		err := invoker(ctx, method, req, reply, cc, opts...)

		// Signal that the message event has been completely processed by the server
		gem.nextEvt(nil, target)

		return err
	}
}
