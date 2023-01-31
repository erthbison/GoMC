package main

import (
	"context"
	"errors"
	"log"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
)

type MessageScheduler struct {
	sync.Mutex
	// msgQueue list.List
	msgQueue []chanMsg

	inChan chan chanMsg
	// connChanMap map[connKey]chan<- struct{}
	signalChan chan struct{}
}

func NewMessageScheduler() *MessageScheduler {
	return &MessageScheduler{
		// msgQueue:    *list.New(),
		msgQueue: []chanMsg{},
		inChan:   make(chan chanMsg),
		// connChanMap: make(map[connKey]chan<- struct{}),
		signalChan: make(chan struct{}),
	}
}

func (msc *MessageScheduler) ListenForMsg() {
	go func() {
		for {
			msg := <-msc.inChan
			msc.Lock()
			msc.msgQueue = append(msc.msgQueue, msg)
			msc.Unlock()
		}
	}()
}

func (msc *MessageScheduler) NextMessage() error {
	msc.Lock()
	defer msc.Unlock()
	if len(msc.msgQueue) == 0 {
		return errors.New("No messages in queue")
	}
	msc.getNextMsg()
	// signalChan := msc.connChanMap[connKey{localId: msg.from, remoteId: msg.to}]
	msc.signalChan <- struct{}{}
	return nil
}

func (msc *MessageScheduler) getNextMsg() chanMsg {
	// Always pick the message with highest sender ID

	tempMsg := msc.msgQueue[0]
	index := 0

	if len(msc.msgQueue) == 1 {
		return tempMsg
	}
	for i := 1; i < len(msc.msgQueue); i++ {
		msg := msc.msgQueue[i]
		num1, _ := strconv.Atoi(msg.from)
		num2, _ := strconv.Atoi(tempMsg.from)
		if num1 > num2 {
			tempMsg = msg
			index = i
		}
	}
	msc.msgQueue = append(msc.msgQueue[:index], msc.msgQueue[index+1:]...)
	return tempMsg
}

func (msc *MessageScheduler) UnaryInterceptWrite(
	ctx context.Context,
	method string,
	req interface{},
	reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	// Get context timeout
	ctx_deadline, ok := ctx.Deadline()
	duration := time.Until(ctx_deadline)

	log.Println("Received Message", req)
	msc.inChan <- chanMsg{
		from: "",
		to:   "",
		msg:  req,
	}
	log.Print("Awaiting indication to release message\n")
	<-msc.signalChan
	log.Println("Released message", req)
	if ok {
		ctx, _ = context.WithTimeout(context.Background(), duration)
	} else {
		ctx = context.Background()
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

func (msc *MessageScheduler) StreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

	go func() {
		var msg interface{}
		for {
			ss.RecvMsg(msg)
			log.Println("Received Message", msg)
			msc.inChan <- chanMsg{
				from: "",
				to:   "",
				msg:  msg,
			}
			log.Print("Awaiting indication to release message\n")
			<-msc.signalChan
		}
	}()
	return handler(srv, ss)
}

type connKey struct {
	localId  string
	remoteId string
}

type chanMsg struct {
	from string
	to   string
	msg  interface{}
}
