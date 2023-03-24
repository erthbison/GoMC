package main

import "time"

type fd struct {
	alive    map[int]bool
	detected map[int]bool

	id    int
	nodes []int

	duration time.Duration
	sleep    func(time.Duration)
	send     func(int, string, ...any)

	stopped bool
	crashed []int
}

func NewFd(id int, nodes []int, duration time.Duration, send func(int, string, ...any), sleep func(time.Duration)) *fd {
	alive := make(map[int]bool)
	for _, id := range nodes {
		alive[id] = true
	}
	return &fd{
		alive:    alive,
		detected: make(map[int]bool),

		id:    id,
		nodes: nodes,

		duration: duration,
		sleep:    sleep,
		send:     send,

		crashed: make([]int, 0),
	}
}

func (fd *fd) Start() {
	for !fd.stopped {
		for _, id := range fd.nodes {
			tmp := !fd.alive[id]
			_ = tmp
			if !fd.alive[id] && !fd.detected[id] {
				fd.detected[id] = true
				// Send crash signal
				fd.crashed = append(fd.crashed, id)
			}
			fd.send(id, "HeartBeatRequest", fd.id)
		}
		fd.alive = make(map[int]bool)
		if len(fd.crashed) == 3 {
			return
		}
		fd.sleep(fd.duration)
	}
}

func (fd *fd) HeartBeatRequest(from int) {
	if fd.stopped {
		return
	}
	fd.send(from, "HeartBeatReply", fd.id)
}

func (fd *fd) HeartBeatReply(from int) {
	if fd.stopped {
		return
	}
	fd.alive[from] = true
}
