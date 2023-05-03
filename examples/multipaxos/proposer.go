package multipaxos

import (
	"context"
	"gomc/examples/multipaxos/proto"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
)

type proposer struct {
	proto.UnimplementedProposerServer

	sync.Mutex

	nodeId int64
	qourum int

	Adu      int64
	nextSlot int64

	crnd int64
	cval *proto.Value

	completedPhase1 bool

	leader *LeaderElector

	promises map[int64]*proto.PromiseRequest

	pendingProposals []string

	nodes       map[int64]*multipaxosClient
	waitForSend func(num int)
	wait        func(time.Duration)
}

func newProposer(id int64, waitForSend func(num int), l *LeaderElector) *proposer {
	p := &proposer{
		nodeId: id,

		crnd: id,

		leader: l,

		pendingProposals: make([]string, 0),

		promises:    make(map[int64]*proto.PromiseRequest),
		waitForSend: waitForSend,
	}
	l.LeaderSubscribe(p.newLeader)
	return p
}

func (p *proposer) performPhaseOne() {
	msg := &proto.PrepareRequest{
		Crnd: p.crnd,
		Slot: p.Adu,
		From: p.nodeId,
	}

	for _, n := range p.nodes {
		go n.Prepare(context.Background(), msg)
	}
	p.waitForSend(len(p.nodes))
}

func (p *proposer) Propose(_ context.Context, prop *proto.ProposeRequest) (*empty.Empty, error) {
	p.Lock()
	defer p.Unlock()

	if !p.leader.IsLeader() {
		return &empty.Empty{}, nil
	}

	if !p.completedPhase1 {
		// Store this value for when phase 1 is completed
		p.pendingProposals = append(p.pendingProposals, prop.GetVal())
		return &empty.Empty{}, nil
	}

	p.nextSlot++
	// create accept message
	acc := &proto.AcceptRequest{
		Val: &proto.Value{
			Val: prop.GetVal(),
			Rnd: p.crnd,
		},
		Slot: p.nextSlot,
		From: p.nodeId,
	}

	// Send accept message
	for _, n := range p.nodes {
		go n.Accept(context.Background(), acc)
	}
	p.waitForSend(len(p.nodes))

	return &empty.Empty{}, nil
}

func (p *proposer) Promise(_ context.Context, prm *proto.PromiseRequest) (*empty.Empty, error) {
	p.Lock()
	defer p.Unlock()

	// Wrong round: ignore
	if prm.GetRnd() != p.crnd {
		return &empty.Empty{}, nil
	}

	// We have already received a promise from this node: ignore
	if _, ok := p.promises[prm.GetFrom()]; ok {
		return &empty.Empty{}, nil
	}

	// We have already completed phase1, so we just ignore this message
	if p.completedPhase1 {
		return &empty.Empty{}, nil
	}

	p.promises[prm.GetFrom()] = prm

	// Still have not reached qourum size
	if len(p.promises) < p.qourum {
		return &empty.Empty{}, nil
	}

	accepts := p.getValues()

	// Accept messages are added to the front of the queue
	// p.valQueue = append(accepts, p.valQueue...)
	for _, val := range accepts {
		p.nextSlot++
		// create accept message
		acc := &proto.AcceptRequest{
			Val: &proto.Value{
				Val: val,
				Rnd: p.crnd,
			},
			Slot: p.nextSlot,
			From: p.nodeId,
		}

		// Send accept message
		for _, n := range p.nodes {
			go n.Accept(context.Background(), acc)
		}
		p.waitForSend(len(p.nodes))
	}

	for _, val := range p.pendingProposals {
		p.nextSlot++
		// create accept message
		acc := &proto.AcceptRequest{
			Val: &proto.Value{
				Val: val,
				Rnd: p.crnd,
			},
			Slot: p.nextSlot,
			From: p.nodeId,
		}

		// Send accept message
		for _, n := range p.nodes {
			go n.Accept(context.Background(), acc)
		}
		p.waitForSend(len(p.nodes))
	}

	p.pendingProposals = make([]string, 0)

	p.completedPhase1 = true
	p.promises = make(map[int64]*proto.PromiseRequest)

	return &empty.Empty{}, nil
}

func (p *proposer) getValues() []string {
	slots := make(map[int64]*proto.PromiseSlot)
	highestSlot := p.Adu
	for _, prm := range p.promises {
		for _, slot := range prm.Slots {
			if slot.GetSlot() > highestSlot {
				highestSlot = slot.GetSlot()
			}

			oldSlot, ok := slots[slot.GetSlot()]
			if !ok {
				slots[slot.GetSlot()] = oldSlot
				continue
			}

			if slot.GetVal().GetRnd() > oldSlot.GetVal().GetRnd() {
				slots[slot.GetSlot()] = slot
			}
		}
	}

	if len(slots) == 0 {
		return []string{}
	}

	length := highestSlot - p.Adu
	vals := make([]string, length)
	for i := 0; i < int(length); i++ {
		slot := slots[p.Adu+int64(i)]
		vals[i] = slot.GetVal().GetVal()
	}
	return vals
}

func (p *proposer) ProposeVal(val string) {
	// Send value to leader
	go p.nodes[p.leader.Leader()].Propose(
		context.Background(),
		&proto.ProposeRequest{Val: val},
	)
	p.waitForSend(1)
}

func (p *proposer) newLeader(newLeader int64) {
	p.Lock()
	defer p.Unlock()

	p.completedPhase1 = false
	p.IncrementCrnd()

	if p.leader.IsLeader() {
		p.performPhaseOne()
	}
}

func (p *proposer) IncrementCrnd() {
	p.crnd += int64(len(p.nodes))
}

func (p *proposer) IncrementAdu() {
	p.Adu++
	if p.Adu > p.nextSlot {
		p.nextSlot = p.Adu
	}
}
