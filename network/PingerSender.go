package main

import (
	"fmt"
	"net"
	"sync"
)

type PingerServer struct {
	connList []net.Conn
	buffsize int
}

func NewPingerServer(connList []net.Conn) *PingerServer {
	return &PingerServer{
		connList: connList,
		buffsize: 1024,
	}
}

func (ps *PingerServer) Ping() {
	wait := sync.WaitGroup{}
	for i := 0; i < len(ps.connList); i++ {
		wait.Add(1)
		go func(conn net.Conn) {
			_, err := conn.Write([]byte("Ping from server"))
			if err != nil {
				panic(err)
			}
			response := make([]byte, ps.buffsize)
			n, err := conn.Read(response)
			if err != nil {
				panic(err)
			}
			fmt.Printf("Received response: %s", response[:n])
			wait.Done()
		}(ps.connList[i])
	}
	wait.Wait()
}

func (ps *PingerServer) PingFirst() {
	conn := ps.connList[0]
	_, err := conn.Write([]byte("Ping from server"))
	if err != nil {
		panic(err)
	}
	response := make([]byte, ps.buffsize)
	n, err := conn.Read(response)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Received response: %s", response[:n])

}
