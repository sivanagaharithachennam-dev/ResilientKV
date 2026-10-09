package raft

import "testing"

func TestStartElection(t *testing.T) {
	node := NewNode("node1")

	term := node.StartElection()

	if term != 1 {
		t.Fatalf("expected term 1, got %d", term)
	}
	if node.State != Candidate {
		t.Fatalf("expected Candidate, got %s", node.State)
	}
	if node.VotedFor != "node1" {
		t.Fatalf("expected self-vote for node1, got %s", node.VotedFor)
	}
}

func TestStartElectionIncrementsTerm(t *testing.T) {
	node := NewNode("node1")

	first := node.StartElection()
	second := node.StartElection()

	if second != first+1 {
		t.Fatalf("expected term to increase by one: first=%d second=%d", first, second)
	}
}

func TestCandidateCanBecomeLeaderForCurrentTerm(t *testing.T) {
	node := NewNode("node1")
	term := node.StartElection()

	if !node.BecomeLeader(term) {
		t.Fatal("expected candidate to become leader for current term")
	}
	if node.State != Leader {
		t.Fatalf("expected Leader, got %s", node.State)
	}
}

func TestCannotBecomeLeaderForOldTerm(t *testing.T) {
	node := NewNode("node1")
	term := node.StartElection()

	if node.BecomeLeader(term - 1) {
		t.Fatal("should not become leader for an old term")
	}
	if node.State != Candidate {
		t.Fatalf("expected Candidate, got %s", node.State)
	}
}

func TestBecomeFollowerOnNewerTerm(t *testing.T) {
	node := NewNode("node1")
	node.StartElection()

	if !node.BecomeFollower(5) {
		t.Fatal("expected transition to Follower")
	}
	if node.CurrentTerm != 5 {
		t.Fatalf("expected term 5, got %d", node.CurrentTerm)
	}
	if node.State != Follower {
		t.Fatalf("expected Follower, got %s", node.State)
	}
	if node.VotedFor != "" {
		t.Fatalf("expected vote reset, got %s", node.VotedFor)
	}
}
