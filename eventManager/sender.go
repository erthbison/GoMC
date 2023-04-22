package eventManager

import (
	"gomc/event"
)

type Sender struct {
	ea EventAdder
}

func NewSender(ea EventAdder) *Sender {
	return &Sender{ea: ea}
}

func (s *Sender) SendFunc(id int) func(int, string, ...any) {
	return func(to int, msgType string, params ...any) {
		s.ea.AddEvent(event.NewMessageHandlerEvent(id, to, msgType, params...))
	}
}
