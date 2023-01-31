package main

import (
	"errors"
	"net"
	"strconv"
	"sync"
)

type MessageScheduler struct {
	sync.Mutex
	// msgQueue list.List
	msgQueue []chanMsg

	inChan      chan chanMsg
	connChanMap map[connKey]chan<- struct{}
}

func NewMessageScheduler() *MessageScheduler {
	return &MessageScheduler{
		// msgQueue:    *list.New(),
		msgQueue:    []chanMsg{},
		inChan:      make(chan chanMsg),
		connChanMap: make(map[connKey]chan<- struct{}),
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
	msg := msc.getNextMsg()
	signalChan := msc.connChanMap[connKey{localId: msg.from, remoteId: msg.to}]
	signalChan <- struct{}{}
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

func (msc *MessageScheduler) NewConn(id1, id2 string) (net.Conn, net.Conn) {
	conn1, conn2, conn1SignalChan, conn2SignalChan := newMockConnPair(id1, id2, msc.inChan)
	msc.connChanMap[connKey{localId: conn1.localId, remoteId: conn1.remoteId}] = conn1SignalChan
	msc.connChanMap[connKey{localId: conn2.localId, remoteId: conn2.remoteId}] = conn2SignalChan
	return conn1, conn2
}

type connKey struct {
	localId  string
	remoteId string
}

type chanMsg struct {
	from string
	to   string
	msg  []byte
}
