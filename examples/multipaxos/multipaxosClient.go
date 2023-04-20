package multipaxos

import (
	"gomc/examples/multipaxos/proto"

	"google.golang.org/grpc"
)

type multipaxosClient struct {
	conn *grpc.ClientConn
	proto.ProposerClient
	proto.AcceptorClient
	proto.LearnerClient
}

func NewMultipaxosClient(conn *grpc.ClientConn) *multipaxosClient {
	return &multipaxosClient{
		conn:           conn,
		ProposerClient: proto.NewProposerClient(conn),
		AcceptorClient: proto.NewAcceptorClient(conn),
		LearnerClient:  proto.NewLearnerClient(conn),
	}
}

func (mpc *multipaxosClient) Stop() {
	mpc.conn.Close()
}
