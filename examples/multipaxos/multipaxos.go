package multipaxos

import (
	"net"

	"google.golang.org/grpc"
)

type MultiPaxos struct {
	Id           int64
	LearntValues map[int64]string

	srv *server

	proposed []string
	stopped  bool

	nodes   map[int64]*multipaxosClient
	addrMap map[int64]string

	*server
	*LeaderElector
	*proposer
	*acceptor
	*learner
}

func NewMultiPaxos(id int64, addrMap map[int64]string, waitForSend func(int)) *MultiPaxos {
	lrn := newLearner(id)
	l := NewLeaderElector(id, addrMap)
	mp := &MultiPaxos{
		Id: id,

		srv: &server{},

		proposed:     make([]string, 0),
		LearntValues: make(map[int64]string),
		addrMap:      addrMap,

		LeaderElector: l,
		proposer:      newProposer(id, waitForSend, l),
		acceptor:      newAcceptor(id, waitForSend),
		learner:       lrn,
	}

	lrn.LearnSubscribe(mp.LearnValue)
	return mp
}

func (mp *MultiPaxos) Start(lis net.Listener, srvOpts ...grpc.ServerOption) {
	go mp.srv.StartServer(mp, lis, srvOpts...)
}

func (mp *MultiPaxos) DialNodes(dialOpts ...grpc.DialOption) error {
	nodes, err := mp.srv.DialNodes(mp.addrMap, dialOpts...)
	if err != nil {
		return err
	}
	qourum := len(nodes)/2 + 1
	mp.nodes = nodes
	mp.proposer.nodes = nodes
	mp.acceptor.nodes = nodes
	mp.proposer.qourum = qourum
	mp.learner.qourum = qourum

	return nil
}

func (mp *MultiPaxos) ProposeVal(val string) {
	if mp.stopped {
		return
	}

	mp.proposed = append(mp.proposed, val)
	mp.proposer.ProposeVal(val)
}

func (mp *MultiPaxos) LearnValue(val string, slot int64) {
	mp.LearntValues[slot] = val
	mp.proposer.IncrementAdu()
}

func (mp *MultiPaxos) Stop() {
	mp.srv.Stop()
	mp.stopped = true

	for _, n := range mp.nodes {
		n.Stop()
	}
}
