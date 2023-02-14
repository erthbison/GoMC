package main

type fifo struct {
	id int

	send func(int, string, ...any)

	Received []string
}

func NewFIFO(id int) *fifo {
	return &fifo{
		id: id,
	}
}

func (f fifo) Send(target int, message []byte) {
	f.send(
		target,
		"Test",
		message,
	)
}

func (f fifo) Test(message []byte) {
	f.Received = append(f.Received, string(message))
}
