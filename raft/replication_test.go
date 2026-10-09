package raft

import (
	"context"
	"net"
	"testing"
	"time"

	pb "resilientkv/grpc/proto"

	"google.golang.org/grpc"
)

type replicationTestServer struct {
	pb.UnimplementedKeyValueServiceServer
	received chan *pb.AppendEntriesRequest
}

func (s *replicationTestServer) AppendEntries(
	ctx context.Context,
	req *pb.AppendEntriesRequest,
) (*pb.AppendEntriesResponse, error) {
	s.received <- req
	return &pb.AppendEntriesResponse{
		Term:    req.GetTerm(),
		Success: true,
	}, nil
}

func TestReplicateLogSendsEntriesOverGRPC(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	mock := &replicationTestServer{
		received: make(chan *pb.AppendEntriesRequest, 1),
	}

	server := grpc.NewServer()
	pb.RegisterKeyValueServiceServer(server, mock)
	go func() {
		_ = server.Serve(listener)
	}()
	defer server.Stop()

	node := NewNode("node1")
	term := node.StartElection()
	if !node.BecomeLeader(term) {
		t.Fatal("expected node to become leader")
	}

	commands := []string{
		`{"op":"put","key":"fruit","value":"apple"}`,
		`{"op":"delete","key":"old"}`,
	}
	for _, command := range commands {
		if _, _, err := node.AppendLeaderEntry(command); err != nil {
			t.Fatal(err)
		}
	}

	acks, err := node.ReplicateLog(
		[]string{listener.Addr().String()},
		time.Second,
	)
	if err != nil {
		t.Fatal(err)
	}
	if acks != 2 {
		t.Fatalf("expected leader plus one follower acknowledgement, got %d", acks)
	}

	select {
	case req := <-mock.received:
		if req.GetLeaderId() != "node1" {
			t.Fatalf("unexpected leader ID: %s", req.GetLeaderId())
		}
		if len(req.GetEntries()) != len(commands) {
			t.Fatalf("expected %d entries, got %d", len(commands), len(req.GetEntries()))
		}
		for i, command := range commands {
			if req.GetEntries()[i].GetCommand() != command {
				t.Fatalf("entry %d mismatch: got %q", i, req.GetEntries()[i].GetCommand())
			}
		}
		t.Log("Follower received both log entries over gRPC")

	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for replicated entries")
	}
}

func TestFollowerCannotReplicateLog(t *testing.T) {
	node := NewNode("node2")

	if _, err := node.ReplicateLog([]string{"127.0.0.1:50052"}, time.Second); err == nil {
		t.Fatal("expected follower replication to fail")
	}
}
