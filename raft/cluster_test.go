package raft

import "testing"

func TestConductElectionWithMajority(t *testing.T) {
	n1 := NewNode("node1")
	n2 := NewNode("node2")
	n3 := NewNode("node3")

	nodes := []*Node{n1, n2, n3}

	if !ConductElection(nodes, n1) {
		t.Fatal("expected node1 to win with a majority")
	}

	if n1.State != Leader {
		t.Fatalf("expected node1 to be Leader, got %s", n1.State)
	}

	if n2.VotedFor != "node1" || n3.VotedFor != "node1" {
		t.Fatal("expected the other nodes to vote for node1")
	}
}

func TestElectionStepsDownWhenHigherTermExists(t *testing.T) {
	n1 := NewNode("node1")
	n2 := NewNode("node2")
	n3 := NewNode("node3")

	// Simulate a node that has already observed a newer term.
	n2.UpdateTerm(5)

	nodes := []*Node{n1, n2, n3}

	if ConductElection(nodes, n1) {
		t.Fatal("candidate must not win after discovering a higher term")
	}

	if n1.State != Follower {
		t.Fatalf("expected candidate to step down to Follower, got %s", n1.State)
	}

	if n1.CurrentTerm != 5 {
		t.Fatalf("expected term 5 after stepping down, got %d", n1.CurrentTerm)
	}
}

func TestElectionRejectsCandidateOutsideCluster(t *testing.T) {
	n1 := NewNode("node1")
	n2 := NewNode("node2")
	outside := NewNode("outside")

	if ConductElection([]*Node{n1, n2}, outside) {
		t.Fatal("expected election to reject a candidate outside the cluster")
	}
}
