package gomc

import (
	"bytes"
	"encoding/gob"
)

type Sender[T any] struct {
	sch Scheduler[T]
}

func NewSender[T any](sch Scheduler[T]) *Sender[T] {
	return &Sender[T]{sch: sch}
}

func (s *Sender[T]) Send(from, to int, msgType string, msg []byte) {
	var buffer bytes.Buffer
	err := gob.NewEncoder(&buffer).Encode(msg)
	if err != nil {
		// Unoptimal error handling, but we do it for simplicity
		panic(err)
	}
	s.sch.AddEvent(NewMessageEvent[T](from, to, msgType, buffer.Bytes()))
}

func (s *Sender[T]) SendInterface(from, to int, msgType string, msg interface{}) {
	var buffer bytes.Buffer
	err := gob.NewEncoder(&buffer).Encode(msg)
	if err != nil {
		// Unoptimal error handling, but we do it for simplicity
		panic(err)
	}
	s.sch.AddEvent(NewMessageEvent[T](from, to, msgType, buffer.Bytes()))
}
