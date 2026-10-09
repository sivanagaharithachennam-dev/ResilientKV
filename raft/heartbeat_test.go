package raft

import (
	"context"
	"net"
	"testing"
	"time"

	pb "resilientkv/grpc/proto"

	"google.golang.org/grpc"
)

type heartbeatTestServer struct {
	pb.UnimplementedKeyValueServiceServer
	received chan *pb.AppendEntriesRequest
}

func (s *heartbeatTestServer) AppendEntries(
	ctx context.Context,
	req *pb.AppendEntriesRequest,
) (*pb.AppendEntriesResponse, error) {
	s.received <- req
	return &pb.AppendEntriesResponse{
		Term:    req.GetTerm(),
		Success: true,
	}, nil
}

func TestStartHeartbeatsSendsGRPCAppendEntries(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	mock := &heartbeatTestServer{
		received: make(chan *pb.AppendEntriesRequest, 4),
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
		t.Fatal("node could not become leader")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop, err := node.StartHeartbeats(
		ctx,
		[]string{listener.Addr().String()},
		HeartbeatConfig{
			Interval: 50 * time.Millisecond,
			Timeout:  time.Second,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer stop()

	select {
	case req := <-mock.received:
		if req.GetLeaderId() != "node1" {
			t.Fatalf("expected leader node1, got %s", req.GetLeaderId())
		}
		if req.GetTerm() != int32(term) {
			t.Fatalf("expected term %d, got %d", term, req.GetTerm())
		}
		t.Log("gRPC heartbeat received from leader node1")

	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for gRPC heartbeat")
	}
}
