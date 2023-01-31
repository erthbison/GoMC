package main

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

func main() {
	// Benchmark()
	Experiment()
}

func Benchmark() {
	numClients := 3
	network := "tcp"

	connSlice := []net.Conn{}
	for i := 0; i < numClients; i++ {
		pinger := NewPingerListener(i)

		lis, err := net.Listen(network, "")
		if err != nil {
			panic(err)
		}
		go func() {
			conn, err := lis.Accept()
			if err != nil {
				panic(err)
			}
			pinger.ListenOnConnection(conn)
		}()
		conn, err := net.Dial(network, lis.Addr().String())
		if err != nil {
			panic(err)
		}
		connSlice = append(connSlice, conn)
	}

	pingerServer := NewPingerServer(connSlice)
	pingerServer.Ping()

	time.Sleep(time.Second * 3)
}

func Experiment() {
	numClients := 3
	ms := NewMessageScheduler()
	ms.ListenForMsg()
	test := sync.WaitGroup{}
	test.Add(1)

	go func() {
		test.Wait()
		for {
			err := ms.NextMessage()
			if err != nil {
				fmt.Println("Received error:", err)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	connSlice := []net.Conn{}

	for i := 0; i < numClients; i++ {
		pinger := NewPingerListener(i)
		lisConn, srvConn := ms.NewConn(strconv.Itoa(i), "-1")

		pinger.ListenOnConnection(lisConn)

		connSlice = append(connSlice, srvConn)
	}

	pingerServer := NewPingerServer(connSlice)

	test.Done()
	pingerServer.Ping()
	// runtime.Gosched()

	// go pingerServer.PingFirst()

}
