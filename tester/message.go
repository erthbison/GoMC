package tester

import (
	"bytes"
	"fmt"
)

type Message struct {
	From  int
	To    int
	Type  string
	Value []byte
}

func (m Message) Equals(b Message) bool {
	if m.From != b.From {
		return false
	}
	if m.To != b.To {
		return false
	}
	if m.Type != b.Type {
		return false
	}
	if !bytes.Equal(m.Value, b.Value) {
		return false
	}
	return true
}

func (m Message) String() string {
	return fmt.Sprintf("{From: %v, To: %v, Type: %s}", m.From, m.To, m.Type)
}
