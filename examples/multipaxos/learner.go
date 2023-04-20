package multipaxos

import (
	"context"
	"gomc/examples/multipaxos/proto"
	"sync"

	"github.com/golang/protobuf/ptypes/empty"
)

type learntSlots struct {
	learnt bool
	rnd    int64
	votes  map[int64]*proto.Value
}

type learner struct {
	proto.UnimplementedLearnerServer

	sync.Mutex

	nodeId int64

	learns map[int64]*learntSlots

	learnSubscribe []func(string, int64)

	qourum int
}

func newLearner(id int64) *learner {
	return &learner{
		nodeId: id,

		learns: make(map[int64]*learntSlots),
	}
}

func (l *learner) Learn(_ context.Context, lrn *proto.LearnRequest) (*empty.Empty, error) {
	l.Lock()
	defer l.Unlock()

	slotId := lrn.GetSlot()
	slot, ok := l.learns[slotId]
	if !ok {
		slot = &learntSlots{}
		l.learns[slotId] = slot
	}

	if slot.learnt {
		return &empty.Empty{}, nil
	}

	// If the round of the stored slot is higher than the round of the received slot.
	if lrn.GetVal().GetRnd() < slot.rnd {
		return &empty.Empty{}, nil
	}

	// The new learn is larger then the current. Set the current to the new and reset votes
	if lrn.GetVal().GetRnd() > slot.rnd {
		slot.rnd = lrn.GetVal().GetRnd()
		slot.votes = make(map[int64]*proto.Value)
	}

	// The learn is for the current round. Add it to the slot
	if lrn.GetVal().GetRnd() == slot.rnd {
		// We have already received a learn from this node for this slot and this round.
		// Ignore this one
		if _, ok := slot.votes[lrn.GetFrom()]; ok {
			return &empty.Empty{}, nil
		}

		slot.votes[lrn.GetFrom()] = lrn.GetVal()
		if len(slot.votes) >= l.qourum {
			slot.learnt = true
			l.emmitLearn(lrn.GetVal(), slotId)
		}
	}
	return &empty.Empty{}, nil
}

func (l *learner) emmitLearn(val *proto.Value, slotId int64) {
	for _, callback := range l.learnSubscribe {
		callback(val.GetVal(), slotId)
	}
}

func (l *learner) LearnSubscribe(callback func(string, int64)) {
	l.learnSubscribe = append(l.learnSubscribe, callback)
}
