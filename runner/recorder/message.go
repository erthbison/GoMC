package recorder

import "fmt"

type Message struct {
	From, To int
	Sent     bool
	Msg      interface{}
}

func (m Message) String() string {
	var t string
	if m.Sent {
		t = "Sent"
	} else {
		t = "Received"
	}
	return fmt.Sprintf("\t[%v From: %v To: %v Type: %T Msg: %v]\n", t, m.From, m.To, m.Msg, m.Msg)
}
