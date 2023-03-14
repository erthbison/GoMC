package main

import (
	"context"
	"fmt"
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

	DecidedVal  []string
	ProposedVal string

	srv *grpc.Server
}

func NewGrpcConsensus(id int32, lis net.Listener, srvOpts ...grpc.ServerOption) *GrpcConsensus {
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

		srv: srv,
	}
	pb.RegisterConsensusServer(srv, gc)
	go func() {
		srv.Serve(lis)
	}()
	return gc
}

func (gc *GrpcConsensus) DialServers(ids []int32, addrMap map[int32]string, dialOpts ...grpc.DialOption) {
	for _, id := range ids {
		addr := addrMap[id]
		conn, err := grpc.Dial(addr, dialOpts...)
		grpc.Dial(addr)
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

func (gc *GrpcConsensus) Crash(id int) {
	gc.detectedRanks[int32(id)] = true
	for gc.delivered[gc.round] || gc.detectedRanks[gc.round] {
		gc.round++
		gc.decide()
	}
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
	if gc.id == gc.round && !gc.broadcast && gc.proposal != nil {
		gc.broadcast = true
		gc.DecidedVal = append(gc.DecidedVal, gc.proposal.GetVal())
		for _, node := range gc.nodes {
			if node.id > gc.id {
				_, err := node.Decided(context.Background(), &pb.DecideRequest{
					Val:  gc.proposal,
					From: gc.id,
				})
				if err != nil {
					fmt.Printf("An error occurred while dialing node %v: %v\n", node.id, err)
				}
			}
		}
	}
}

func (gc *GrpcConsensus) Stop() {
	gc.srv.Stop()
	for _, node := range gc.nodes {
		node.conn.Close()
	}
}
