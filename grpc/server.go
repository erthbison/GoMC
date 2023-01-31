package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"sync"

	pb "experimentation/grpc/proto"
)

type keyValueServicesServer struct {
	kv map[string]string
	// this must be included in implementers of the pb.KeyValueServicesServer interface
	pb.UnimplementedKeyValueServiceServer
	kvMutex sync.RWMutex
}

// NewKeyValueServicesServer returns an initialized KeyValueServicesServer
func NewKeyValueServicesServer() *keyValueServicesServer {
	return &keyValueServicesServer{
		kv: make(map[string]string),
	}
}

var (
	help = flag.Bool(
		"help",
		false,
		"Show usage help",
	)
	endpoint = flag.String(
		"endpoint",
		"localhost:12111",
		"Endpoint on which server runs or to which client connects",
	)
)

// Usage prints usage info
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
}

// **************************************************************************************************************
// The Insert() gRPC inserts a key/value pair into the map.
// Input:  ctx     The context of the client's request.
//
//	req     The request from the client. Contains a key/value pair.
//
// Output: (1)     A response to the client containing whether or not the insert was successful.
//
//	(2)     An error (if any).
//
// **************************************************************************************************************
func (s *keyValueServicesServer) Insert(ctx context.Context, req *pb.InsertRequest) (*pb.InsertResponse, error) {
	s.kvMutex.Lock()
	s.kv[req.Key] = req.Value
	s.kvMutex.Unlock()
	return &pb.InsertResponse{Success: true}, nil
}

// **************************************************************************************************************
// The Lookup() gRPC returns a value corresponding to the key provided in the input.
// Input:  ctx     The context of the client's request.
//
//	req     The request from the client. Contains a key pair.
//
// Output: (1)     A response to the client containing the value corresponding to the key.
//
//	(2)     An error (if any).
//
// **************************************************************************************************************
func (s *keyValueServicesServer) Lookup(ctx context.Context, req *pb.LookupRequest) (*pb.LookupResponse, error) {
	s.kvMutex.RLock()
	if value, ok := s.kv[req.Key]; ok {
		return &pb.LookupResponse{Value: value}, nil
	}
	s.kvMutex.RUnlock()
	return &pb.LookupResponse{Value: ""}, nil
}

// **************************************************************************************************************
// The Keys() gRPC returns a slice listing all the keys.
// Input:  ctx     The context of the client's request.
//
//	req     The request from the client.
//
// Output: (1)     A response to the client containing a slice of the keys.
//
//	(2)     An error (if any).
//
// **************************************************************************************************************
func (s *keyValueServicesServer) Keys(ctx context.Context, req *pb.KeysRequest) (*pb.KeysResponse, error) {
	s.kvMutex.RLock()
	keySlice := []string{}
	for key := range s.kv {
		keySlice = append(keySlice, key)
	}
	s.kvMutex.RUnlock()
	return &pb.KeysResponse{Keys: keySlice}, nil
}

func (s *keyValueServicesServer) InsertStream(stream pb.KeyValueService_InsertStreamServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		s.kvMutex.Lock()
		s.kv[in.Key] = in.Value
		s.kvMutex.Unlock()

		stream.Send(
			&pb.InsertResponse{
				Success: true,
			},
		)
	}
}

// func main() {
// 	flag.Usage = Usage
// 	flag.Parse()
// 	if *help {
// 		flag.Usage()
// 		return
// 	}

// 	listener, err := net.Listen("tcp", *endpoint)
// 	if err != nil {
// 		fmt.Printf("Error: %v\n", err)
// 	} else {
// 		fmt.Printf("Listener started on %v\n", *endpoint)
// 	}

// 	server := NewKeyValueServicesServer()
// 	server.kv = make(map[string]string)
// 	grpcServer := grpc.NewServer()
// 	pb.RegisterKeyValueServiceServer(grpcServer, server)
// 	fmt.Printf("Preparing to serve incoming requests.\n")
// 	err = grpcServer.Serve(listener)
// 	if err != nil {
// 		fmt.Printf("Error: %v\n", err)
// 	}
// }
