package raft_test

import (
	"net"
	"path/filepath"
	"testing"
	"time"

	rpcserver "resilientkv/grpc"
	pb "resilientkv/grpc/proto"
	"resilientkv/raft"
	"resilientkv/storage"

	"google.golang.org/grpc"
)

func TestReplicationToRealRaftFollower(t *testing.T) {
	// Create a real follower storage engine and Raft node.
	engine, err := storage.New(filepath.Join(t.TempDir(), "follower-data"))
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()

	follower := raft.NewNode("node2")
	followerService := rpcserver.NewServerWithRaft(engine, follower)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterKeyValueServiceServer(grpcServer, followerService)
	go func() {
		_ = grpcServer.Serve(listener)
	}()
	defer grpcServer.Stop()

	// Create a leader and append an actual log command.
	leader := raft.NewNode("node1")
	term := leader.StartElection()
	if !leader.BecomeLeader(term) {
		t.Fatal("expected node1 to become leader")
	}

	command := `{"op":"put","key":"fruit","value":"apple"}`
	if _, _, err := leader.AppendLeaderEntry(command); err != nil {
		t.Fatal(err)
	}

	// Send the leader's log to the real follower over gRPC.
	acks, err := leader.ReplicateLog(
		[]string{listener.Addr().String()},
		2*time.Second,
	)
	if err != nil {
		t.Fatal(err)
	}
	if acks != 2 {
		t.Fatalf("expected 2 acknowledgements, got %d", acks)
	}

	// Verify that the follower actually recorded the entry.
	_, state, entries, _ := follower.LogSnapshot()
	if state != raft.Follower {
		t.Fatalf("expected follower state, got %s", state)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 follower log entry, got %d", len(entries))
	}
	if entries[0].Command != command {
		t.Fatalf("unexpected replicated command: %q", entries[0].Command)
	}

	// Log replication alone must not apply the command to storage.
	if _, found := engine.Get("fruit"); found {
		t.Fatal("uncommitted entry must not be applied to storage")
	}

	t.Log("Real follower received the log entry; storage remains unchanged before commit")
}
