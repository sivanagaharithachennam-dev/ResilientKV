package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "resilientkv/grpc/proto"
)

func main() {
	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("connection failed: %v", err)
	}
	defer conn.Close()

	client := pb.NewKeyValueServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	requestID := "put-1001"

	// First request
	log.Println("Sending PUT request...")
	response, err := client.Put(ctx, &pb.PutRequest{
		Key:       "product",
		Value:     "laptop",
		RequestId: requestID,
	})
	if err != nil {
		log.Fatalf("PUT failed: %v", err)
	}

	log.Printf("First PUT: success=%v", response.GetSuccess())

	// Retry SAME request
	log.Println("Retrying SAME PUT request...")
	response, err = client.Put(ctx, &pb.PutRequest{
		Key:       "product",
		Value:     "laptop",
		RequestId: requestID,
	})
	if err != nil {
		log.Fatalf("PUT retry failed: %v", err)
	}

	log.Printf("Retry PUT: success=%v", response.GetSuccess())

	// Verify
	getResponse, err := client.Get(ctx, &pb.GetRequest{
		Key: "product",
	})
	if err != nil {
		log.Fatalf("GET failed: %v", err)
	}

	log.Printf(
		"GET: key=%s value=%s found=%v",
		"product",
		getResponse.GetValue(),
		getResponse.GetFound(),
	)
}
