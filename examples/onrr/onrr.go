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
	readList map[int]Value // A slice of all values

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
	send func(to int, msgType string, params ...any)
}

func NewOnrr(id int, send func(to int, msgType string, params ...any), nodes []int) *onrr {
	return &onrr{
		val:      Value{Ts: 0, Val: 0},
		wts:      0,
		acks:     0,
		rid:      0,
		readList: make(map[int]Value),

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

	// msg := encodeMsg(BroadcastWriteMsg{
	// 	Val: value,
	// })
	msg := BroadcastWriteMsg{Val: value}
	for _, target := range onrr.nodes {
		onrr.send(target, "BroadcastWrite", onrr.id, msg)
	}
}

func (onrr *onrr) BroadcastWrite(from int, msg BroadcastWriteMsg) {
	// bwMsg := decodeMsg[BroadcastWriteMsg](msg)
	bwMsg := msg
	if bwMsg.Val.Ts > onrr.val.Ts {
		onrr.val = bwMsg.Val
	}
	ackMsg := AckMsg{
		Ts: bwMsg.Val.Ts,
	}
	onrr.send(from, "AckWrite", ackMsg)
}

func (onrr *onrr) AckWrite(msg AckMsg) {
	// ackMsg := decodeMsg[AckMsg](msg)
	ackMsg := msg
	if ackMsg.Ts != onrr.wts {
		return
	}
	onrr.acks++
	if onrr.acks > len(onrr.nodes)/2 {
		if onrr.ongoingWrite {
			onrr.possibleReads = onrr.possibleReads[1:]
		}
		onrr.ongoingWrite = false
		onrr.acks = 0
		onrr.WriteIndicator <- true
	}
}

func (onrr *onrr) Read() {
	onrr.ongoingRead = true

	onrr.rid++
	onrr.readList = make(map[int]Value)
	msg := BroadcastReadMsg{
		Rid: onrr.rid,
	}
	for _, target := range onrr.nodes {
		onrr.send(target, "BroadcastRead", onrr.id, msg)
	}
}

func (onrr *onrr) BroadcastRead(from int, msg BroadcastReadMsg) {
	// readMsg := decodeMsg[BroadcastReadMsg](msg)
	readMsg := msg
	valMsg := ReadValueMsg{
		Rid: readMsg.Rid,
		Val: onrr.val,
	}
	onrr.send(from, "ReadValue", onrr.id, valMsg)
}

func (onrr *onrr) ReadValue(from int, msg ReadValueMsg) {
	// valMsg := decodeMsg[ReadValueMsg](msg)
	valMsg := msg
	if valMsg.Rid != onrr.rid {
		return
	}
	onrr.readList[from] = valMsg.Val
	if len(onrr.readList) > len(onrr.nodes)/2 {
		val := getvalue(onrr.readList)
		onrr.readList = make(map[int]Value)

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
		// Un-optimal error handling, but we do it for simplicity
		panic(err)
	}
	return buffer.Bytes()
}

func decodeMsg[T any](input []byte) T {
	var msg T
	buffer := bytes.NewBuffer(input)
	err := gob.NewDecoder(buffer).Decode(&msg)
	if err != nil {
		// Un-optimal error handling, but we do it for simplicity
		panic(err)
	}
	return msg
}
