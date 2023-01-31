package main

import (
	"net"
	"time"
)

type mockConn struct {
	localId  string
	remoteId string

	writeSignal  chan struct{}
	writeMsgChan chan<- chanMsg

	RxRead <-chan []byte
	TxRead chan<- int

	RxWrite <-chan int
	TxWrite chan<- []byte
}

func newMockConnPair(id1, id2 string, writeChan chan chanMsg) (mockConn, mockConn, chan<- struct{}, chan<- struct{}) {
	bc1 := make(chan []byte)
	nc1 := make(chan int)

	bc2 := make(chan []byte)
	nc2 := make(chan int)

	sc1 := make(chan struct{})
	sc2 := make(chan struct{})
	return mockConn{
			localId:  id1,
			remoteId: id2,

			writeMsgChan: writeChan,
			writeSignal:  sc1,

			// channels used to send messages to other channel in pair
			RxRead:  bc1,
			TxRead:  nc1,
			RxWrite: nc2,
			TxWrite: bc2,
		}, mockConn{
			localId:  id2,
			remoteId: id1,

			writeMsgChan: writeChan,
			writeSignal:  sc2,

			// channels used to send messages to other channel in pair
			RxRead:  bc2,
			TxRead:  nc2,
			RxWrite: nc1,
			TxWrite: bc1,
		}, sc1, sc2
}

// Read reads data from the connection.
// Read can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetReadDeadline.
func (mc mockConn) Read(b []byte) (n int, err error) {
	msg := <-mc.RxRead
	n = copy(b, msg)
	mc.TxRead <- n
	return n, nil
}

// Write writes data to the connection.
// Write can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetWriteDeadline.
func (mc mockConn) Write(b []byte) (n int, err error) {
	msg := make([]byte, 1024)
	size := copy(msg, b)

	// Before sending the message over the channel:
	// send it to the message scheduler and wait for an indication that we can send
	mc.writeMsgChan <- chanMsg{
		from: mc.localId,
		to:   mc.remoteId,
		msg:  msg[:size],
	}
	<-mc.writeSignal

	mc.TxWrite <- msg[:size]
	n = <-mc.RxWrite
	return n, nil
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (mc mockConn) Close() error {
	close(mc.TxRead)
	close(mc.TxWrite)
	return nil
}

// LocalAddr returns the local network address, if known.
func (mc mockConn) LocalAddr() net.Addr {
	panic("not implemented") // TODO: Implement
}

// RemoteAddr returns the remote network address, if known.
func (mc mockConn) RemoteAddr() net.Addr {
	panic("not implemented") // TODO: Implement
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail instead of blocking. The deadline applies to all future
// and pending I/O, not just the immediately following call to
// Read or Write. After a deadline has been exceeded, the
// connection can be refreshed by setting a deadline in the future.
//
// If the deadline is exceeded a call to Read or Write or to other
// I/O methods will return an error that wraps os.ErrDeadlineExceeded.
// This can be tested using errors.Is(err, os.ErrDeadlineExceeded).
// The error's Timeout method will return true, but note that there
// are other possible errors for which the Timeout method will
// return true even if the deadline has not been exceeded.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (mc mockConn) SetDeadline(t time.Time) error {
	panic("not implemented") // TODO: Implement
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (mc mockConn) SetReadDeadline(t time.Time) error {
	panic("not implemented") // TODO: Implement
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (mc mockConn) SetWriteDeadline(t time.Time) error {
	panic("not implemented") // TODO: Implement
}
