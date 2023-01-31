package main

import (
	"fmt"
	"io"
	"net"
)

type PingerListener struct {
	id       int
	buffsize int
}

func NewPingerListener(id int) *PingerListener {
	return &PingerListener{
		id:       id,
		buffsize: 1024,
	}
}
func (p *PingerListener) ListenOnConnection(conn net.Conn) {
	go func(conn net.Conn) {
		for {
			buffer := make([]byte, p.buffsize)
			n, err := conn.Read(buffer)
			if err != nil {
				switch err {
				case io.EOF:
					return
				default:
					panic(err)
				}
			}
			if n > p.buffsize {
				panic("Too large message received")
			}
			response := fmt.Sprintf("%d: Ping received: %q\n", p.id, string(buffer[:n]))
			fmt.Print(response)
			_, err = conn.Write([]byte(response))
			if err != nil {
				switch err {
				case io.EOF:
					return
				default:
					panic(err)
				}
			}
		}
	}(conn)
}
