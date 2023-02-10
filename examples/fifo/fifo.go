package main

type fifo struct {
	id int

	send func(int, string, []byte)

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

func (f fifo) Test(from, to int, message []byte) {
	f.Received = append(f.Received, string(message))
}
