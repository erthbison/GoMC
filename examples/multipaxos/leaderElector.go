package multipaxos

type LeaderElector struct {
	id      int64
	correct map[int64]bool
	leader  int64

	leaderSub []func(int64)

	stopped bool
}

func NewLeaderElector(id int64, addr map[int64]string) *LeaderElector {
	var leader int64
	correct := make(map[int64]bool)
	for id := range addr {
		correct[id] = true
		if id > leader {
			leader = id
		}
	}
	return &LeaderElector{
		id:      id,
		correct: correct,
		leader:  leader,
	}
}

func (l *LeaderElector) NodeCrash(id int, _ bool) {
	if l.stopped {
		return
	}

	l.correct[int64(id)] = false

	if int64(id) == l.leader {
		leader := l.nextLeader()
		l.leader = leader
		for _, f := range l.leaderSub {
			f(leader)
		}
	}
}

func (l *LeaderElector) nextLeader() int64 {
	var leader int64
	for id, ok := range l.correct {
		if ok {
			if id > leader {
				leader = id
			}
		}
	}
	return leader
}

// Subscribe to leader change calls
func (l *LeaderElector) LeaderSubscribe(f func(id int64)) {
	l.leaderSub = append(l.leaderSub, f)
}

func (l *LeaderElector) IsLeader() bool {
	return l.leader == l.id
}

func (l *LeaderElector) Leader() int64 {
	return l.leader
}
