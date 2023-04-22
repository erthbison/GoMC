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

	numNodes int

	// Consensus Value
	learnSubscribe []func(string)
}

func NewLearner(id *proto.NodeId, numNodes int) *Learner {
	return &Learner{
		id: id,

		numNodes: numNodes,
		recvLrn:  make(map[int64]*proto.Value),

		learnSubscribe: make([]func(string), 0),
	}
}

func (l *Learner) Learn(_ context.Context, in *proto.LearnRequest) (*empty.Empty, error) {
	// l.Lock()
	// defer l.Unlock()

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
	if numLrn > l.numNodes/2 {
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

func (l *Learner) Subscribe(f func(val string)) {
	l.learnSubscribe = append(l.learnSubscribe, f)
}

func (l *Learner) emmitLearn(val *proto.Value) {
	for _, f := range l.learnSubscribe {
		f(val.GetVal())
	}
}
