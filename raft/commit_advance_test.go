package raft

import "testing"

func newCommitTestLeader(t *testing.T) *Node {
	t.Helper()

	n := NewNode("node1")
	term := n.StartElection()
	if !n.BecomeLeader(term) {
		t.Fatal("expected node to become leader")
	}
	return n
}

func TestAdvanceCommitWithMajority(t *testing.T) {
	n := newCommitTestLeader(t)

	if _, _, err := n.AppendLeaderEntry("command-1"); err != nil {
		t.Fatal(err)
	}

	advanced, err := n.AdvanceCommitWithMajority(1, 2, 3)
	if err != nil {
		t.Fatal(err)
	}
	if !advanced {
		t.Fatal("expected majority to advance commit index")
	}

	_, _, _, commitIndex := n.LogSnapshot()
	if commitIndex != 1 {
		t.Fatalf("commit index = %d, want 1", commitIndex)
	}
}

func TestAdvanceCommitWithoutMajority(t *testing.T) {
	n := newCommitTestLeader(t)
	if _, _, err := n.AppendLeaderEntry("command-1"); err != nil {
		t.Fatal(err)
	}

	advanced, err := n.AdvanceCommitWithMajority(1, 1, 3)
	if err != nil {
		t.Fatal(err)
	}
	if advanced {
		t.Fatal("commit index advanced without a majority")
	}

	_, _, _, commitIndex := n.LogSnapshot()
	if commitIndex != 0 {
		t.Fatalf("commit index = %d, want 0", commitIndex)
	}
}

func TestAdvanceCommitRequiresCurrentTermEntry(t *testing.T) {
	n := newCommitTestLeader(t)

	// Simulate a previous-term entry in the log.
	n.mu.Lock()
	n.Log = append(n.Log, LogEntry{Term: n.CurrentTerm - 1, Command: "old"})
	n.mu.Unlock()

	advanced, err := n.AdvanceCommitWithMajority(1, 3, 3)
	if err != nil {
		t.Fatal(err)
	}
	if advanced {
		t.Fatal("should not directly commit an entry from an older term")
	}
}

func TestAdvanceCommitRejectsFollower(t *testing.T) {
	n := NewNode("node1")

	advanced, err := n.AdvanceCommitWithMajority(1, 1, 1)
	if err == nil {
		t.Fatal("expected follower to be rejected")
	}
	if advanced {
		t.Fatal("follower must not advance commit index")
	}
}

func TestAdvanceCommitRejectsInvalidAcknowledgements(t *testing.T) {
	n := newCommitTestLeader(t)
	if _, _, err := n.AppendLeaderEntry("command-1"); err != nil {
		t.Fatal(err)
	}

	if _, err := n.AdvanceCommitWithMajority(1, 4, 3); err == nil {
		t.Fatal("expected excessive acknowledgement count to be rejected")
	}
}
