package raft

import "testing"

func TestAppendEntriesReplacesConflictingUncommittedSuffix(t *testing.T) {
	n := NewNode("follower")
	n.CurrentTerm = 2
	n.Log = []LogEntry{
		{Term: 1, Command: "first"},
		{Term: 1, Command: "old"},
		{Term: 2, Command: "old-tail"},
	}
	n.CommitIndex = 1

	resp, err := n.HandleAppendEntries(AppendEntriesRequest{
		Term:         2,
		LeaderID:     "leader",
		PrevLogIndex: 1,
		PrevLogTerm:  1,
		Entries: []LogEntry{
			{Term: 2, Command: "replacement"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Success {
		t.Fatal("expected conflicting uncommitted suffix to be replaced")
	}
	if len(n.Log) != 2 || n.Log[1].Command != "replacement" {
		t.Fatalf("unexpected reconciled log: %+v", n.Log)
	}
}

func TestAppendEntriesRejectsConflictWithCommittedEntry(t *testing.T) {
	n := NewNode("follower")
	n.CurrentTerm = 2
	n.Log = []LogEntry{
		{Term: 1, Command: "first"},
		{Term: 1, Command: "committed"},
	}
	n.CommitIndex = 2

	resp, err := n.HandleAppendEntries(AppendEntriesRequest{
		Term:         2,
		LeaderID:     "leader",
		PrevLogIndex: 1,
		PrevLogTerm:  1,
		Entries: []LogEntry{
			{Term: 2, Command: "replacement"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Success {
		t.Fatal("expected conflict with a committed entry to be rejected")
	}
	if n.Log[1].Command != "committed" {
		t.Fatal("committed entry was changed")
	}
}

func TestAppendEntriesHeartbeatPreservesExtraSuffix(t *testing.T) {
	n := NewNode("follower")
	n.CurrentTerm = 2
	n.Log = []LogEntry{
		{Term: 1, Command: "first"},
		{Term: 2, Command: "existing-suffix"},
	}

	resp, err := n.HandleAppendEntries(AppendEntriesRequest{
		Term:         2,
		LeaderID:     "leader",
		PrevLogIndex: 1,
		PrevLogTerm:  1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Success {
		t.Fatal("expected heartbeat to succeed")
	}
	if len(n.Log) != 2 || n.Log[1].Command != "existing-suffix" {
		t.Fatalf("heartbeat unexpectedly changed the log: %+v", n.Log)
	}
}

func TestAppendEntriesRejectsDifferentCommandWithSameTerm(t *testing.T) {
	n := NewNode("follower")
	n.CurrentTerm = 2
	n.Log = []LogEntry{
		{Term: 1, Command: "first"},
		{Term: 2, Command: "original"},
	}

	resp, err := n.HandleAppendEntries(AppendEntriesRequest{
		Term:         2,
		LeaderID:     "leader",
		PrevLogIndex: 1,
		PrevLogTerm:  1,
		Entries: []LogEntry{
			{Term: 2, Command: "different"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Success {
		t.Fatal("expected inconsistent same-term entry to be rejected")
	}
	if n.Log[1].Command != "original" {
		t.Fatal("existing entry was unexpectedly changed")
	}
}
