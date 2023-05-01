package main

import (
	"context"
	pb "gomc/examples/grpcConsensus/proto"
	"net"

	"google.golang.org/grpc"
)

type state struct {
	proposed string
	decided  []string
}

type Client struct {
	pb.ConsensusClient
	id   int32
	conn *grpc.ClientConn
}

type GrpcConsensus struct {
	pb.UnimplementedConsensusServer

	detectedRanks map[int32]bool
	proposal      *pb.Value
	proposer      int32

	round     int32
	delivered map[int32]bool
	broadcast bool

	id    int32
	nodes []*Client

	stopped bool

	DecidedVal  []string
	ProposedVal string

	waitForSend func(num int)

	srv *grpc.Server
}

func NewGrpcConsensus(id int32, lis net.Listener, waitForSend func(int), srvOpts ...grpc.ServerOption) *GrpcConsensus {
	srv := grpc.NewServer(srvOpts...)
	gc := &GrpcConsensus{
		detectedRanks: make(map[int32]bool),
		proposer:      0,
		round:         1,
		delivered:     make(map[int32]bool),
		broadcast:     false,
		id:            id,
		nodes:         make([]*Client, 0),

		DecidedVal:  make([]string, 0),
		ProposedVal: "",

		waitForSend: waitForSend,

		srv: srv,
	}
	pb.RegisterConsensusServer(srv, gc)
	go func() {
		srv.Serve(lis)
	}()
	return gc
}

func (gc *GrpcConsensus) DialServers(addrMap map[int32]string, dialOpts ...grpc.DialOption) {
	for id, addr := range addrMap {
		conn, err := grpc.Dial(addr, dialOpts...)
		if err != nil {
			panic(err)
		}
		gc.nodes = append(gc.nodes, &Client{
			id:              id,
			ConsensusClient: pb.NewConsensusClient(conn),
			conn:            conn,
		})
	}
}

func (gc *GrpcConsensus) Crash(id int, _ bool) {
	if gc.stopped {
		return
	}
	gc.detectedRanks[int32(id)] = true
	// for gc.delivered[gc.round] || gc.detectedRanks[gc.round] {
	// 	gc.round++
	// 	gc.decide()
	// }
}

func (hc *GrpcConsensus) Propose(val string) {
	hc.ProposedVal = val
	protoVal := &pb.Value{Val: val}
	if hc.proposal == nil {
		hc.proposal = protoVal
	}
	hc.decide()
}

func (gc *GrpcConsensus) Decided(ctx context.Context, in *pb.DecideRequest) (*pb.DecideResponse, error) {
	if in.GetFrom() < gc.id && in.GetFrom() > gc.proposer {
		gc.proposal = in.GetVal()
		gc.proposer = in.GetFrom()
		gc.decide()
	}
	gc.delivered[in.GetFrom()] = true
	for gc.delivered[gc.round] || gc.detectedRanks[gc.round] {
		gc.round++
		gc.decide()
	}
	return &pb.DecideResponse{}, nil
}

func (gc *GrpcConsensus) decide() {
	if gc.id != gc.round {
		return
	}

	if gc.broadcast {
		return
	}
	if gc.proposal == nil {
		return
	}

	gc.broadcast = true
	gc.DecidedVal = append(gc.DecidedVal, gc.proposal.GetVal())
	msg := &pb.DecideRequest{
		Val:  gc.proposal,
		From: gc.id,
	}
	num := 0
	for _, node := range gc.nodes {
		if node.id > gc.id {
			num++
			go node.Decided(context.Background(), msg)
		}
	}
	gc.waitForSend(num) // Wait until all messages has been sent, but do not wait until an answer is received
}

func (gc *GrpcConsensus) Stop() {
	gc.srv.Stop()
	for _, node := range gc.nodes {
		node.conn.Close()
	}
	gc.stopped = true
}
