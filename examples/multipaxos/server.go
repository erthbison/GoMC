package multipaxos

import (
	"gomc/examples/multipaxos/proto"
	"net"

	"google.golang.org/grpc"
)

type server struct {
	srv   *grpc.Server
	conns []*grpc.ClientConn
}

func (s *server) StartServer(mp *MultiPaxos, lis net.Listener, srvOpts ...grpc.ServerOption) error {
	s.srv = grpc.NewServer(srvOpts...)
	proto.RegisterProposerServer(s.srv, mp)
	proto.RegisterAcceptorServer(s.srv, mp)
	proto.RegisterLearnerServer(s.srv, mp)
	return s.srv.Serve(lis)
}

func (s *server) Stop() {
	s.srv.Stop()
	for _, c := range s.conns {
		c.Close()
	}
}

func (s *server) DialNodes(addrMap map[int64]string, dialOpts ...grpc.DialOption) (map[int64]*multipaxosClient, error) {
	nodes := make(map[int64]*multipaxosClient)
	for id, addr := range addrMap {
		conn, err := grpc.Dial(addr, dialOpts...)
		if err != nil {
			return nil, err
		}
		nodes[id] = NewMultipaxosClient(conn)
	}
	return nodes, nil
}
