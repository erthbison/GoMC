package main

import (
	"bytes"
	"encoding/gob"
)

type Value struct {
	Ts  int
	Val int
}

type BroadcastWriteMsg struct {
	Val Value
}

type AckMsg struct {
	Ts int
}

type BroadcastReadMsg struct {
	Rid int
}

type ReadValueMsg struct {
	Rid int
	Val Value
}

type onrr struct {
	val      Value         // Current value stored in the register
	wts      int           // Write timestamp
	acks     int           // Number of acks received for the current value
	rid      int           // A read request identifier
	readlist map[int]Value // A slice of all values

	WriteIndicator chan bool
	ReadIndicator  chan int

	// Used for testing.
	ongoingRead   bool
	ongoingWrite  bool
	possibleReads []int

	// Id of the current node
	id int
	// Used to keep track of all nodes
	nodes []int
	// Used to send messages to other types.
	send func(from, to int, msgType string, msg []byte)
}

func NewOnrr(id int, send func(from, to int, msgType string, msg []byte), nodes []int) *onrr {
	return &onrr{
		val:      Value{Ts: 0, Val: 0},
		wts:      0,
		acks:     0,
		rid:      0,
		readlist: make(map[int]Value),

		// Indicator channels
		WriteIndicator: make(chan bool, 1),
		ReadIndicator:  make(chan int, 1),

		ongoingRead:   false,
		ongoingWrite:  false,
		possibleReads: []int{0},

		id:    id,
		nodes: nodes,
		send:  send,
	}
}

func (onrr *onrr) Write(val int) {
	onrr.wts++
	onrr.acks = 0

	onrr.possibleReads = []int{onrr.val.Val, val}

	value := Value{
		Ts:  onrr.wts,
		Val: val,
	}

	onrr.ongoingWrite = true

	msg := encodeMsg(
		BroadcastWriteMsg{
			Val: value,
		},
	)
	for _, target := range onrr.nodes {
		onrr.send(onrr.id, target, "BroadcastWrite", msg)
	}
}

func (onrr *onrr) BroadcastWrite(from int, to int, msg []byte) {
	bwMsg := decodeMsg[BroadcastWriteMsg](msg)
	if bwMsg.Val.Ts > onrr.val.Ts {
		onrr.val = bwMsg.Val
	}
	ackMsg := encodeMsg(AckMsg{
		Ts: bwMsg.Val.Ts,
	})
	onrr.send(onrr.id, from, "AckWrite", ackMsg)
}

func (onrr *onrr) AckWrite(from int, to int, msg []byte) {
	ackMsg := decodeMsg[AckMsg](msg)
	if ackMsg.Ts != onrr.wts {
		return
	}
	onrr.acks++
	if onrr.acks > len(onrr.nodes)/2 {
		onrr.ongoingWrite = false
		onrr.possibleReads = []int{onrr.val.Val}
		onrr.acks = 0
		onrr.WriteIndicator <- true
	}
}

func (onrr *onrr) Read() {
	onrr.ongoingRead = true

	onrr.rid++
	onrr.readlist = make(map[int]Value)
	msg := encodeMsg(BroadcastReadMsg{
		Rid: onrr.rid,
	})
	for _, target := range onrr.nodes {
		onrr.send(onrr.id, target, "BroadcastRead", msg)
	}
}

func (onrr *onrr) BroadcastRead(from int, to int, msg []byte) {
	readMsg := decodeMsg[BroadcastReadMsg](msg)
	valMsg := encodeMsg(ReadValueMsg{
		Rid: readMsg.Rid,
		Val: onrr.val,
	})
	onrr.send(onrr.id, from, "ReadValue", valMsg)
}

func (onrr *onrr) ReadValue(from int, to int, msg []byte) {
	valMsg := decodeMsg[ReadValueMsg](msg)
	if valMsg.Rid != onrr.rid {
		return
	}
	onrr.readlist[from] = valMsg.Val
	if len(onrr.readlist) > len(onrr.nodes)/2 {
		val := getvalue(onrr.readlist)
		onrr.readlist = make(map[int]Value)

		onrr.ongoingRead = false
		onrr.ReadIndicator <- val.Val
	}
}

func getvalue(valueMap map[int]Value) Value {
	var highest Value
	for _, val := range valueMap {
		if val.Ts > highest.Ts {
			highest = val
		}
	}
	return highest
}

func encodeMsg(v interface{}) []byte {
	var buffer bytes.Buffer
	err := gob.NewEncoder(&buffer).Encode(v)
	if err != nil {
		// Unoptimal error handling, but we do it for simplicity
		panic(err)
	}
	return buffer.Bytes()
}

func decodeMsg[T any](input []byte) T {
	var msg T
	buffer := bytes.NewBuffer(input)
	err := gob.NewDecoder(buffer).Decode(&msg)
	if err != nil {
		// Unoptimal error handling, but we do it for simplicity
		panic(err)
	}
	return msg
}