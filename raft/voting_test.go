package raft

import "testing"

func TestGrantFirstVote(t *testing.T) {
	node := NewNode("node1")

	if !node.RequestVote("candidate1", 1, 0, 0) {
		t.Fatal("expected first vote to be granted")
	}
	if node.CurrentTerm != 1 {
		t.Fatalf("expected term 1, got %d", node.CurrentTerm)
	}
	if node.VotedFor != "candidate1" {
		t.Fatalf("expected vote for candidate1, got %s", node.VotedFor)
	}
}

func TestRejectSecondCandidateInSameTerm(t *testing.T) {
	node := NewNode("node1")

	if !node.RequestVote("candidate1", 1, 0, 0) {
		t.Fatal("expected first vote to be granted")
	}
	if node.RequestVote("candidate2", 1, 0, 0) {
		t.Fatal("should not grant a second vote in the same term")
	}
}

func TestAllowSameCandidateRetry(t *testing.T) {
	node := NewNode("node1")

	if !node.RequestVote("candidate1", 1, 0, 0) {
		t.Fatal("expected first vote to be granted")
	}
	if !node.RequestVote("candidate1", 1, 0, 0) {
		t.Fatal("repeat request from the same candidate should be accepted")
	}
}

func TestRejectOlderTerm(t *testing.T) {
	node := NewNode("node1")
	node.UpdateTerm(3)

	if node.RequestVote("candidate1", 2, 0, 0) {
		t.Fatal("should reject a candidate from an older term")
	}
	if node.CurrentTerm != 3 {
		t.Fatalf("expected term to remain 3, got %d", node.CurrentTerm)
	}
}

func TestNewerTermResetsVote(t *testing.T) {
	node := NewNode("node1")

	if !node.RequestVote("candidate1", 1, 0, 0) {
		t.Fatal("expected first vote to be granted")
	}
	node.UpdateTerm(2)

	if node.CurrentTerm != 2 || node.VotedFor != "" || node.State != Follower {
		t.Fatal("newer term should reset the vote and return to Follower")
	}
}

func TestRejectCandidateWithOutdatedLog(t *testing.T) {
	node := NewNode("node1")
	node.Log = []LogEntry{
		{Term: 2, Command: "put a=1"},
		{Term: 3, Command: "put b=2"},
	}
	node.UpdateTerm(3)

	// Candidate's log ends in term 2, while this node's log ends in term 3.
	if node.RequestVote("candidate1", 4, 2, 2) {
		t.Fatal("should reject a candidate with an outdated log")
	}
}

func TestGrantVoteToCandidateWithNewerLog(t *testing.T) {
	node := NewNode("node1")
	node.Log = []LogEntry{{Term: 2, Command: "put a=1"}}
	node.UpdateTerm(2)

	if !node.RequestVote("candidate1", 3, 2, 3) {
		t.Fatal("expected vote for candidate with a newer log term")
	}
}
