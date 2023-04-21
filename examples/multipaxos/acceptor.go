package multipaxos

import (
	"context"
	"gomc/examples/multipaxos/proto"
	"sync"

	"github.com/golang/protobuf/ptypes/empty"
)

type acceptor struct {
	proto.UnimplementedAcceptorServer

	sync.Mutex

	nodeId int64

	rnd int64

	slots map[int64]*proto.PromiseSlot

	nodes       map[int64]*multipaxosClient
	waitForSend func(num int)
}

func newAcceptor(id int64, waitForSend func(num int)) *acceptor {
	return &acceptor{
		nodeId: id,

		slots:       make(map[int64]*proto.PromiseSlot),
		waitForSend: waitForSend,
	}
}

func (a *acceptor) Prepare(_ context.Context, prp *proto.PrepareRequest) (*empty.Empty, error) {
	a.Lock()
	defer a.Unlock()

	if prp.GetCrnd() <= a.rnd {
		return &empty.Empty{}, nil
	}

	a.rnd = prp.GetCrnd()

	slots := make([]*proto.PromiseSlot, 0)
	for _, slot := range a.slots {
		if slot.Slot >= prp.Slot {
			slots = append(slots, slot)
		}
	}

	go a.nodes[prp.GetFrom()].Promise(
		context.Background(),
		&proto.PromiseRequest{
			Rnd:   a.rnd,
			Slots: slots,
			From:  a.nodeId,
		},
	)
	a.waitForSend(1)

	return &empty.Empty{}, nil
}

func (a *acceptor) Accept(_ context.Context, acc *proto.AcceptRequest) (*empty.Empty, error) {
	a.Lock()
	defer a.Unlock()

	if acc.GetVal().GetRnd() < a.rnd {
		return &empty.Empty{}, nil
	}

	a.rnd = acc.GetVal().GetRnd()
	a.slots[acc.GetSlot()] = &proto.PromiseSlot{
		Slot: acc.GetSlot(),
		Val:  acc.GetVal(),
	}

	lrn := &proto.LearnRequest{
		Val:  acc.GetVal(),
		Slot: acc.GetSlot(),
		From: a.nodeId,
	}

	for _, n := range a.nodes {
		go n.Learn(context.Background(), lrn)
	}
	a.waitForSend(len(a.nodes))

	return &empty.Empty{}, nil
}
