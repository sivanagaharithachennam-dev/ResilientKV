package raft

import "testing"

func TestNewNodeStartsAsFollower(t *testing.T) {
	node := NewNode("node1")

	if node.ID != "node1" {
		t.Fatalf("expected node ID node1, got %s", node.ID)
	}

	if node.State != Follower {
		t.Fatalf("expected Follower state, got %s", node.State)
	}

	if node.CurrentTerm != 0 {
		t.Fatalf("expected initial term 0, got %d", node.CurrentTerm)
	}

	if node.VotedFor != "" {
		t.Fatalf("expected no initial vote, got %s", node.VotedFor)
	}

	if node.CommitIndex != 0 || node.LastApplied != 0 {
		t.Fatal("expected initial commit index and last applied index to be 0")
	}
}

func TestRaftStateNames(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{Follower, "Follower"},
		{Candidate, "Candidate"},
		{Leader, "Leader"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("State.String() = %q, want %q", got, tt.want)
		}
	}
}
