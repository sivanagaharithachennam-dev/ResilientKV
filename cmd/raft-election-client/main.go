package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	pb "resilientkv/grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	target := "localhost:50051"
	if value := os.Getenv("RAFT_TARGET"); value != "" {
		target = value
	}

	peersValue := os.Getenv("RAFT_PEERS")
	if peersValue == "" {
		peersValue = "localhost:50052,localhost:50053"
	}

	var peers []string
	for _, address := range strings.Split(peersValue, ",") {
		address = strings.TrimSpace(address)
		if address != "" {
			peers = append(peers, address)
		}
	}
	if len(peers) == 0 {
		log.Fatal("RAFT_PEERS must contain at least one address")
	}

	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("connect to %s: %v", target, err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := pb.NewKeyValueServiceClient(conn).TriggerElection(
		ctx,
		&pb.TriggerElectionRequest{
			PeerAddresses: peers,
			TimeoutMillis: 2000,
		},
	)
	if err != nil {
		log.Fatalf("trigger election: %v", err)
	}

	log.Printf(
		"Election result: elected=%t term=%d state=%s",
		response.GetElected(),
		response.GetTerm(),
		response.GetState(),
	)
}
