package eventManager

import (
	"gomc/event"
	"gomc/scheduler"
)

type Sender struct {
	sch scheduler.Scheduler
}

func NewSender(sch scheduler.Scheduler) *Sender {
	return &Sender{sch: sch}
}

func (s *Sender) SendFunc(id int) func(int, string, ...any) {
	return func(to int, msgType string, params ...any) {
		s.sch.AddEvent(event.NewMessageEvent(id, to, msgType, params...))
	}
}
