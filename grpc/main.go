package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	pb "experimentation/grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	endpoint := "localhost:12112"

	msc := NewMessageScheduler()
	msc.ListenForMsg()
	go startServer(endpoint, msc)

	go func() {
		input := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("Press enter to release next message\n")
			input.Scan()
			err := msc.NextMessage()
			if err != nil {
				fmt.Println("An error occurred:", err)
			}
		}
	}()
	runClient(endpoint, msc)
}

func runClient(endpoint string, msc *MessageScheduler) {

	conn, err := grpc.Dial(
		endpoint,
		grpc.WithUnaryInterceptor(msc.UnaryInterceptWrite),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := pb.NewKeyValueServiceClient(conn)

	testData := map[string]string{"a": "alpha", "b": "bravo", "c": "charlie"}

	// Unary request
	for key, value := range testData {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		resp, err := client.Insert(
			ctx,
			&pb.InsertRequest{Key: key, Value: value})
		cancel()
		if err != nil {
			panic(err)
		}

		log.Println("Received Response:", resp)
	}

	// // Stream request
	// stream, err := client.InsertStream(context.Background())
	// if err != nil {
	// 	panic(err)
	// }
	// for key, value := range testData {
	// 	err := stream.Send(&pb.InsertRequest{
	// 		Key:   key,
	// 		Value: value,
	// 	})
	// 	if err == io.EOF {
	// 		return
	// 	} else if err != nil {
	// 		panic(err)
	// 	}
	// }
	// err = stream.CloseSend()
	// if err != nil {
	// 	panic(err)
	// }

	keysResponse, err := client.Keys(context.Background(), &pb.KeysRequest{})
	if err != nil {
		panic(err)
	}
	keys := keysResponse.Keys
	var numKeys int
	for _, key := range keys {
		if _, ok := testData[key]; ok {
			numKeys += 1
		}
	}
	fmt.Println("Keys:", keysResponse.Keys, "Num Keys:", numKeys)
}

func startServer(endpoint string, msc *MessageScheduler) {

	listener, err := net.Listen("tcp", endpoint)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Listener started on %v\n", endpoint)
	}

	server := NewKeyValueServicesServer()
	server.kv = make(map[string]string)
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(msc.StreamServerInterceptor),
	)

	pb.RegisterKeyValueServiceServer(grpcServer, server)
	fmt.Printf("Preparing to serve incoming requests.\n")
	err = grpcServer.Serve(listener)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}