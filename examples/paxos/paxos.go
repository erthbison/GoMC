package paxos

import (
	"gomc/examples/paxos/proto"
	"net"
	"sync"

	"google.golang.org/grpc"
)

type paxosClient struct {
	conn *grpc.ClientConn

	proto.ProposerClient
	proto.AcceptorClient
	proto.LearnerClient
}

func newPaxosClient(conn *grpc.ClientConn) *paxosClient {
	return &paxosClient{
		ProposerClient: proto.NewProposerClient(conn),
		AcceptorClient: proto.NewAcceptorClient(conn),
		LearnerClient:  proto.NewLearnerClient(conn),
	}
}

type paxos struct {
	*Proposer
	*Acceptor
	*Learner

	Proposal string

	stopped bool

	Id      int64
	correct map[int64]bool
	leader  int64

	Lock *sync.Mutex
}

func newPaxos(id int64, nodes map[int64]string, waitForSend func(int)) *paxos {
	nodeId := &proto.NodeId{Val: id}
	var leader int64
	correct := make(map[int64]bool)
	for nodeId := range nodes {
		if nodeId > leader {
			leader = nodeId
		}
		correct[nodeId] = true
	}
	lock := new(sync.Mutex)
	return &paxos{
		Proposer: NewProposer(nodeId, waitForSend),
		Acceptor: NewAcceptor(nodeId, waitForSend),
		Learner:  NewLearner(nodeId, len(nodes), lock),

		Id:      id,
		leader:  leader,
		correct: correct,

		Lock: lock,
	}
}

func (p *paxos) NodeCrash(id int) {
	if p.stopped {
		return
	}
	p.correct[int64(id)] = false
	if int64(id) == p.leader {
		p.newLeader()
	}
}

func (p *paxos) newLeader() {
	p.Proposer.IncrementCrnd()
	p.leader = p.nextLeader()
	if p.leader == p.Id && p.Proposal != "" {
		p.performPrepare(p.Proposal)
	}
}

func (p *paxos) nextLeader() int64 {
	var leader int64
	for id, ok := range p.correct {
		if ok {
			if id > leader {
				leader = id
			}
		}
	}
	return leader
}

func (p *paxos) Propose(val string) {
	if p.stopped {
		return
	}
	p.Proposal = val
	if p.leader == p.Id {
		p.performPrepare(p.Proposal)
	}
}

type Server struct {
	srv         *grpc.Server
	connections []*grpc.ClientConn
	addrMap     map[int64]string

	*paxos
}

func NewServer(id int64, addrMap map[int64]string, waitForSend func(int), srvOpts ...grpc.ServerOption) (*Server, error) {
	srv := grpc.NewServer(srvOpts...)
	paxos := newPaxos(id, addrMap, waitForSend)
	proto.RegisterProposerServer(srv, paxos)
	proto.RegisterAcceptorServer(srv, paxos)
	proto.RegisterLearnerServer(srv, paxos)
	return &Server{
		srv:     srv,
		paxos:   paxos,
		addrMap: addrMap,
	}, nil
}

func (p *Server) StartServer(lis net.Listener) error {
	return p.srv.Serve(lis)
}

func (p *Server) Stop() {
	p.srv.Stop()
	for _, c := range p.connections {
		c.Close()
	}
	p.Close()
	p.stopped = true
}

func (p *Server) DialNodes(dialOpts ...grpc.DialOption) error {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	nodes := make(map[int64]*paxosClient)
	for id, addr := range p.addrMap {
		conn, err := grpc.Dial(addr, dialOpts...)
		if err != nil {
			return err
		}
		nodes[id] = newPaxosClient(conn)
		p.connections = append(p.connections, conn)
	}
	p.paxos.Proposer.nodes = nodes
	p.paxos.Acceptor.nodes = nodes
	return nil
}
