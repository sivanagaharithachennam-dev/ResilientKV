package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "resilientkv/grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to create gRPC connection: %v", err)
	}
	defer conn.Close()

	client := pb.NewKeyValueServiceClient(conn)

	// First candidate requests a vote in term 1.
	first, err := client.RequestVote(ctx, &pb.RequestVoteRequest{
		CandidateId:  "node2",
		Term:         1,
		LastLogIndex: 0,
		LastLogTerm:  0,
	})
	if err != nil {
		log.Fatalf("first vote request failed: %v", err)
	}

	fmt.Printf(
		"First request: term=%d, vote_granted=%t\n",
		first.GetTerm(),
		first.GetVoteGranted(),
	)

	if !first.GetVoteGranted() {
		log.Fatal("expected node2 to receive the first vote")
	}

	// A different candidate must not receive a second vote
	// from this node in the same term.
	second, err := client.RequestVote(ctx, &pb.RequestVoteRequest{
		CandidateId:  "node3",
		Term:         1,
		LastLogIndex: 0,
		LastLogTerm:  0,
	})
	if err != nil {
		log.Fatalf("second vote request failed: %v", err)
	}

	fmt.Printf(
		"Second request: term=%d, vote_granted=%t\n",
		second.GetTerm(),
		second.GetVoteGranted(),
	)

	if second.GetVoteGranted() {
		log.Fatal("incorrectly granted two different candidates votes in the same term")
	}

	fmt.Println("PASS: single-vote-per-term behavior verified")
}
