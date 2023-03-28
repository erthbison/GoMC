package paxos

import (
	"context"
	"gomc/examples/paxos/proto"
	"sync"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/protobuf/types/known/emptypb"
)

func ValEquals(a, b *proto.Value) bool {
	if a.GetRnd().GetVal() != b.GetRnd().GetVal() {
		return false
	}
	if a.GetVal() != b.GetVal() {
		return false
	}
	return true
}

type Learner struct {
	sync.Mutex
	proto.UnimplementedLearnerServer

	id *proto.NodeId

	recvLrn map[int64]*proto.Value

	// Consensus Value
	val *proto.Value

	nodes       map[int64]*paxosClient
	waitForSend func(id int, num int)

	learnSubscribe []chan string
}

func NewLearner(id *proto.NodeId, waitForSend func(id int, num int)) *Learner {
	return &Learner{
		id: id,

		recvLrn:     make(map[int64]*proto.Value),
		nodes:       make(map[int64]*paxosClient),
		waitForSend: waitForSend,

		learnSubscribe: make([]chan string, 0),
	}
}

func (l *Learner) Learn(_ context.Context, in *proto.LearnRequest) (*empty.Empty, error) {
	l.Lock()
	defer l.Unlock()

	l.addValue(in.GetFrom().GetVal(), in.GetVal())
	numLrn := 0
	freqVal := &proto.Value{
		Rnd: &proto.Round{Val: -1},
	}
	for _, val := range l.recvLrn {
		if val.GetRnd().GetVal() > freqVal.GetRnd().GetVal() {
			freqVal = val
			numLrn = 0
		}
		if ValEquals(freqVal, val) {
			numLrn++
		}
	}
	if numLrn > len(l.nodes)/2 {
		l.emmitLearn(freqVal)
	}
	return &emptypb.Empty{}, nil
}

// Add the value to the received value map, if it is the highest round number received from that node
func (l *Learner) addValue(from int64, val *proto.Value) {
	oldVal, ok := l.recvLrn[from]
	if !ok {
		l.recvLrn[from] = val
		return
	}
	if val.GetRnd().GetVal() > oldVal.GetRnd().GetVal() {
		l.recvLrn[from] = val
	}
}

func (l *Learner) Subscribe() <-chan string {
	c := make(chan string)
	l.learnSubscribe = append(l.learnSubscribe, c)
	return c
}

func (l *Learner) emmitLearn(val *proto.Value) {
	l.val = val
	for _, c := range l.learnSubscribe {
		c <- val.GetVal()
	}
}

func (l *Learner) Close() {
	for _, c := range l.learnSubscribe {
		close(c)
	}
}
