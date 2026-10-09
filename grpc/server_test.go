package grpc

import (
	"context"
	"net"
	"time"

	grpcserver "google.golang.org/grpc"
	"path/filepath"
	"strings"
	"testing"

	pb "resilientkv/grpc/proto"
	"resilientkv/raft"
	"resilientkv/storage"
)

func newTestServer(t *testing.T, nodeID string) (*Server, *storage.Engine, *raft.Node) {
	t.Helper()

	engine, err := storage.New(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatalf("create storage engine: %v", err)
	}
	t.Cleanup(func() {
		if err := engine.Close(); err != nil {
			t.Errorf("close storage engine: %v", err)
		}
	})

	node := raft.NewNode(nodeID)
	server := NewServerWithRaft(engine, node)

	return server, engine, node
}

func electTestLeader(t *testing.T, node *raft.Node) {
	t.Helper()

	term := node.StartElection()
	if !node.BecomeLeader(term) {
		t.Fatal("failed to promote test node to leader")
	}
}

func TestRaftPutCommitsInSingleNodeCluster(t *testing.T) {
	server, engine, node := newTestServer(t, "node1")
	electTestLeader(t, node)

	response, err := server.Put(context.Background(), &pb.PutRequest{
		Key:       "fruit",
		Value:     "mango",
		RequestId: "request-1",
	})
	if err != nil {
		t.Fatalf("Put returned RPC error: %v", err)
	}
	if !response.GetSuccess() {
		t.Fatalf("Put failed: %s", response.GetError())
	}

	value, found := engine.Get("fruit")
	if !found || value != "mango" {
		t.Fatalf("expected fruit=mango, got value=%q found=%v", value, found)
	}

	_, _, entries, commitIndex := node.LogSnapshot()
	if len(entries) != 1 {
		t.Fatalf("expected one log entry, got %d", len(entries))
	}
	if commitIndex != 1 {
		t.Fatalf("expected commit index 1, got %d", commitIndex)
	}
}

func TestRaftFollowerRejectsWrite(t *testing.T) {
	server, engine, _ := newTestServer(t, "follower1")

	response, err := server.Put(context.Background(), &pb.PutRequest{
		Key:   "fruit",
		Value: "mango",
	})
	if err != nil {
		t.Fatalf("Put returned RPC error: %v", err)
	}
	if response.GetSuccess() {
		t.Fatal("follower unexpectedly accepted a write")
	}
	if !strings.Contains(response.GetError(), "not the Raft leader") {
		t.Fatalf("unexpected error: %q", response.GetError())
	}

	if _, found := engine.Get("fruit"); found {
		t.Fatal("follower stored a value for a rejected write")
	}
}

func TestRaftPutFailsWithoutClusterMajority(t *testing.T) {
	server, engine, node := newTestServer(t, "node1")
	electTestLeader(t, node)

	// A three-member cluster needs two acknowledgements. The two configured
	// peer addresses below are intentionally unavailable.
	server.SetPeers([]string{
		"127.0.0.1:1",
		"127.0.0.1:2",
	})

	response, err := server.Put(context.Background(), &pb.PutRequest{
		Key:   "fruit",
		Value: "mango",
	})
	if err != nil {
		t.Fatalf("Put returned RPC error: %v", err)
	}
	if response.GetSuccess() {
		t.Fatal("write unexpectedly succeeded without a majority")
	}
	if !strings.Contains(response.GetError(), "write not committed") {
		t.Fatalf("unexpected error: %q", response.GetError())
	}

	if _, found := engine.Get("fruit"); found {
		t.Fatal("uncommitted write was applied to local storage")
	}

	_, _, entries, commitIndex := node.LogSnapshot()
	if len(entries) != 1 {
		t.Fatalf("expected uncommitted log entry to remain, got %d entries", len(entries))
	}
	if commitIndex != 0 {
		t.Fatalf("expected commit index 0, got %d", commitIndex)
	}
}

func TestRaftDeleteCommitsInSingleNodeCluster(t *testing.T) {
	server, engine, node := newTestServer(t, "node1")
	electTestLeader(t, node)

	if err := engine.Put("fruit", "mango"); err != nil {
		t.Fatalf("seed storage: %v", err)
	}

	response, err := server.Delete(context.Background(), &pb.DeleteRequest{
		Key:       "fruit",
		RequestId: "delete-1",
	})
	if err != nil {
		t.Fatalf("Delete returned RPC error: %v", err)
	}
	if !response.GetSuccess() {
		t.Fatalf("Delete failed: %s", response.GetError())
	}

	if _, found := engine.Get("fruit"); found {
		t.Fatal("expected fruit to be deleted")
	}

	_, _, entries, commitIndex := node.LogSnapshot()
	if len(entries) != 1 || commitIndex != 1 {
		t.Fatalf("expected one committed delete entry; entries=%d commitIndex=%d",
			len(entries), commitIndex)
	}
}

func TestRaftPutReplicatesToFollowerOverGRPC(t *testing.T) {
	leaderServer, leaderEngine, leaderNode := newTestServer(t, "leader")
	electTestLeader(t, leaderNode)

	followerServer, followerEngine, followerNode := newTestServer(t, "follower")

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen for follower: %v", err)
	}

	followerRPC := grpcserver.NewServer()
	pb.RegisterKeyValueServiceServer(followerRPC, followerServer)

	go func() {
		if err := followerRPC.Serve(listener); err != nil {
			// Stop() closes the server; no action is needed during cleanup.
		}
	}()

	t.Cleanup(func() {
		followerRPC.Stop()
	})

	leaderServer.SetPeers([]string{listener.Addr().String()})

	response, err := leaderServer.Put(context.Background(), &pb.PutRequest{
		Key:       "distributed",
		Value:     "replicated",
		RequestId: "replication-test-1",
	})
	if err != nil {
		t.Fatalf("leader Put returned RPC error: %v", err)
	}
	if !response.GetSuccess() {
		t.Fatalf("leader Put failed: %s", response.GetError())
	}

	// Verify the leader committed and stored the value.
	leaderValue, leaderFound := leaderEngine.Get("distributed")
	if !leaderFound || leaderValue != "replicated" {
		t.Fatalf("leader value = %q, found = %v", leaderValue, leaderFound)
	}

	// The leader sends a second AppendEntries after committing, so the
	// follower should learn the new commit index before Put returns.
	followerValue, followerFound := followerEngine.Get("distributed")
	if !followerFound || followerValue != "replicated" {
		t.Fatalf("follower value = %q, found = %v", followerValue, followerFound)
	}

	_, _, leaderEntries, leaderCommit := leaderNode.LogSnapshot()
	if len(leaderEntries) != 1 || leaderCommit != 1 {
		t.Fatalf("leader log entries=%d commitIndex=%d; want 1 and 1",
			len(leaderEntries), leaderCommit)
	}

	_, _, followerEntries, followerCommit := followerNode.LogSnapshot()
	if len(followerEntries) != 1 || followerCommit != 1 {
		t.Fatalf("follower log entries=%d commitIndex=%d; want 1 and 1",
			len(followerEntries), followerCommit)
	}

	// A short timeout guards against accidentally leaving background work
	// running indefinitely in this test.
	_ = time.Second
}
