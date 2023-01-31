package main

import (
	"context"
	"time"

	"google.golang.org/grpc"

	pb "experimentation/grpc/proto"
)

func main() {
	conn, err := grpc.Dial(
		"localhost:12111")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := pb.NewKeyValueServiceClient(conn)

	testData := map[string]string{"a": "alpha", "b": "bravo", "c": "charlie"}
	for key, value := range testData {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_, err := client.Insert(
			ctx,
			&pb.InsertRequest{Key: key, Value: value})
		cancel()
		if err != nil {
			panic(err)
		}
	}

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
}
