package paxos

import (
	"context"
	"gomc/examples/paxos/proto"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Acceptor struct {
	proto.UnimplementedAcceptorServer

	id *proto.NodeId

	// Current round
	rnd *proto.Round

	// The last accepted value and the round in which it was accepted
	vval *proto.Value

	nodes       map[int64]*paxosClient
	waitForSend func(num int)
}

func NewAcceptor(id *proto.NodeId, waitForSend func(num int)) *Acceptor {
	return &Acceptor{
		id: id,

		nodes:       make(map[int64]*paxosClient),
		waitForSend: waitForSend,
	}
}

func (a *Acceptor) Prepare(_ context.Context, in *proto.PrepareRequest) (*empty.Empty, error) {
	if in.GetCrnd().GetVal() > a.rnd.GetVal() {
		a.rnd = in.GetCrnd()
	}

	go a.nodes[in.GetFrom().GetVal()].Promise(
		context.Background(),
		&proto.PromiseRequest{
			Rnd:  a.rnd,
			Val:  a.vval,
			From: a.id,
		},
	)
	a.waitForSend(1)
	return &emptypb.Empty{}, nil
}

func (a *Acceptor) Accept(_ context.Context, in *proto.AcceptRequest) (*empty.Empty, error) {
	if in.GetVal().GetRnd().GetVal() < a.rnd.GetVal() {
		return &emptypb.Empty{}, nil
	}
	a.vval = in.GetVal()

	msg := &proto.LearnRequest{
		Val:  in.GetVal(),
		From: a.id,
	}
	for _, node := range a.nodes {
		go node.Learn(context.Background(), msg)
	}
	a.waitForSend(len(a.nodes))
	return &emptypb.Empty{}, nil
}
