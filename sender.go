package gomc

type Sender[T any] struct {
	sch Scheduler[T]
}

func NewSender[T any](sch Scheduler[T]) *Sender[T] {
	return &Sender[T]{sch: sch}
}

func (s *Sender[T]) Send(from, to int, msgType string, msg []byte) {
	s.sch.AddEvent(NewMessageEvent[T](from, to, msgType, msg))
}
