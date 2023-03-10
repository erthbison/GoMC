package main

import (
	"context"
	"fmt"
	pb "gomc/examples/grpcConsensus/proto"
	"net"

	"google.golang.org/grpc"
)

type GrpcConsensus struct {
	pb.UnimplementedConsensusServer

	detectedRanks map[int32]bool
	proposal      *pb.Value
	proposer      int32

	round     int32
	delivered map[int32]bool
	broadcast bool

	id    int32
	nodes map[int32]pb.ConsensusClient
}

func NewGrpcConsensus(id int32, lis net.Listener) *GrpcConsensus {
	gc := &GrpcConsensus{
		detectedRanks: make(map[int32]bool),
		proposer:      0,
		round:         1,
		delivered:     make(map[int32]bool),
		broadcast:     false,
		id:            id,
		nodes:         make(map[int32]pb.ConsensusClient),
	}
	srv := grpc.NewServer()
	pb.RegisterConsensusServer(srv, gc)
	go func() {
		srv.Serve(lis)
	}()
	return gc
}

func (gc *GrpcConsensus) DialServers(addrMap map[int32]string, dialOpts ...grpc.DialOption) {
	for id, addr := range addrMap {
		conn, err := grpc.Dial(addr, dialOpts...)
		grpc.Dial(addr)
		if err != nil {
			panic(err)
		}
		client := pb.NewConsensusClient(conn)
		gc.nodes[id] = client
	}
}

func (gc *GrpcConsensus) Crash(id int) {
	gc.detectedRanks[int32(id)] = true
	for gc.delivered[gc.round] || gc.detectedRanks[gc.round] {
		gc.round++
		gc.decide()
	}
}

func (hc *GrpcConsensus) Propose(val string) {
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
	if gc.id == gc.round && !gc.broadcast && gc.proposal != nil {
		gc.broadcast = true
		for id, node := range gc.nodes {
			if id > gc.id {
				node.Decided(context.Background(), &pb.DecideRequest{
					Val:  gc.proposal,
					From: gc.id,
				})
			}
		}
		fmt.Println("Decided Value: ", gc.proposal.GetVal())
	}
}
