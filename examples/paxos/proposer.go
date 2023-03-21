package paxos

import (
	"context"
	"gomc/examples/paxos/proto"
	"sync"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Proposer struct {
	sync.Mutex
	proto.UnimplementedProposerServer

	id *proto.NodeId

	v *proto.Value

	// Current round
	crnd *proto.Round
	// Constrained consensus value
	cval *proto.Value

	numPromise int
	largestVal *proto.Value

	phaseOne chan bool

	nodes       map[int64]*paxosClient
	waitForSend func(id int, num int)
}

func NewProposer(id *proto.NodeId, waitForSend func(id int, num int)) *Proposer {
	return &Proposer{
		id: id,

		crnd: &proto.Round{Val: id.GetVal()},

		nodes:       make(map[int64]*paxosClient),
		waitForSend: waitForSend,
	}
}

func (p *Proposer) performPrepare(propsedVal string) {
	p.Lock()
	defer p.Unlock()
	// Use a zero value for round. This will always be smaller than any value returned by the acceptor.
	// This ensures that the value is only chosen if no value is returned by an acceptor
	p.largestVal = &proto.Value{
		Val: propsedVal,
	}

	msg := &proto.PrepareRequest{
		Crnd: p.crnd,
		From: p.id,
	}
	for _, n := range p.nodes {
		go n.Prepare(context.Background(), msg)
	}
	// p.waitForSend(int(p.id.GetVal()), len(p.nodes))
}

func (p *Proposer) Promise(_ context.Context, in *proto.PromiseRequest) (*empty.Empty, error) {
	p.Lock()
	defer p.Unlock()
	if in.GetRnd().GetVal() != p.crnd.GetVal() {
		return &emptypb.Empty{}, nil
	}

	p.numPromise++
	if in.GetVal().GetRnd().GetVal() > p.largestVal.GetRnd().GetVal() {
		p.largestVal = in.GetVal()
	}

	if p.numPromise <= len(p.nodes)/2 {
		return &emptypb.Empty{}, nil
	}

	msg := &proto.AcceptRequest{
		Val: &proto.Value{
			Rnd: p.crnd,
			Val: p.largestVal.GetVal(),
		},
		From: p.id,
	}

	for _, node := range p.nodes {
		go node.Accept(context.Background(), msg)
	}
	p.waitForSend(int(p.id.GetVal()), len(p.nodes))

	p.numPromise = 0
	p.largestVal = nil

	return &emptypb.Empty{}, nil
}

func (p *Proposer) IncrementCrnd() {
	newRnd := p.crnd.GetVal() + int64(len(p.nodes))
	p.crnd = &proto.Round{Val: newRnd}
}
