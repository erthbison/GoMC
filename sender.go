package gomc

import (
	"gomc/event"
	"gomc/scheduler"
)

type Sender[T any] struct {
	sch scheduler.Scheduler[T]
}

func NewSender[T any](sch scheduler.Scheduler[T]) *Sender[T] {
	return &Sender[T]{sch: sch}
}

func (s *Sender[T]) Send(from, to int, msgType string, msg []byte) {
	s.sch.AddEvent(event.NewMessageEvent[T](from, to, msgType, msg))
}
