package main

import (
	"log"
	"net"
	"os"
	"strings"
	"time"

	rpcserver "resilientkv/grpc"
	pb "resilientkv/grpc/proto"
	"resilientkv/raft"
	"resilientkv/storage"

	"google.golang.org/grpc"
)

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func parsePeers(value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}

	var peers []string
	seen := make(map[string]bool)

	for _, item := range strings.Split(value, ",") {
		address := strings.TrimSpace(item)

		if address == "" {
			return nil, os.ErrInvalid
		}
		if seen[address] {
			return nil, os.ErrInvalid
		}

		seen[address] = true
		peers = append(peers, address)
	}

	return peers, nil
}

func main() {
	nodeID := envOrDefault("RAFT_NODE_ID", "node1")
	port := envOrDefault("RESILIENTKV_PORT", "50051")
	dataDir := envOrDefault("RESILIENTKV_DATA_DIR", "./data")

	peers, err := parsePeers(os.Getenv("RAFT_PEERS"))
	if err != nil {
		log.Fatalf("invalid RAFT_PEERS configuration: %v", err)
	}

	if strings.EqualFold(os.Getenv("RAFT_START_ELECTION"), "true") &&
		len(peers) == 0 {
		log.Fatal("RAFT_START_ELECTION=true requires RAFT_PEERS")
	}

	engine, err := storage.New(dataDir)
	if err != nil {
		log.Fatalf("failed to create storage engine: %v", err)
	}
	defer engine.Close()

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", port, err)
	}

	raftNode := raft.NewNode(nodeID)
	grpcServer := grpc.NewServer()

	server := rpcserver.NewServerWithRaft(engine, raftNode)
	server.SetPeers(peers)
	pb.RegisterKeyValueServiceServer(grpcServer, server)

	if strings.EqualFold(os.Getenv("RAFT_START_ELECTION"), "true") {
		log.Printf("Node %s starting network election; peers=%v", nodeID, peers)

		elected, electionErr := raftNode.ConductNetworkElection(
			peers,
			2*time.Second,
		)
		if electionErr != nil {
			log.Printf("Network election failed: %v", electionErr)
		} else if elected {
			log.Printf(
				"Node %s won the election for term %d",
				nodeID,
				raftNode.GetCurrentTerm(),
			)
		} else {
			log.Printf(
				"Node %s did not obtain a majority; state=%s term=%d",
				nodeID,
				raftNode.GetState(),
				raftNode.GetCurrentTerm(),
			)
		}
	}

	log.Printf(
		"ResilientKV server running: node=%s port=%s state=%s peers=%v",
		nodeID,
		port,
		raftNode.GetState(),
		peers,
	)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("gRPC server failed: %v", err)
	}
}
