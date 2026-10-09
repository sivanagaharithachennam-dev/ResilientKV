package raft

import "testing"

func TestLeaderLogAppendCommitAndApply(t *testing.T) {
	node := NewNode("node1")

	term := node.StartElection()
	if !node.BecomeLeader(term) {
		t.Fatal("expected node to become leader")
	}

	index, entryTerm, err := node.AppendLeaderEntry(`{"op":"put","key":"fruit","value":"apple"}`)
	if err != nil {
		t.Fatal(err)
	}
	if index != 1 {
		t.Fatalf("expected entry index 1, got %d", index)
	}
	if entryTerm != term {
		t.Fatalf("expected term %d, got %d", term, entryTerm)
	}

	gotTerm, gotState, entries, commitIndex := node.LogSnapshot()
	if gotTerm != term || gotState != Leader {
		t.Fatalf("unexpected snapshot: term=%d state=%s", gotTerm, gotState)
	}
	if len(entries) != 1 || entries[0].Command == "" {
		t.Fatalf("expected one non-empty log entry, got %#v", entries)
	}
	if commitIndex != 0 {
		t.Fatalf("new entry should not be committed yet; got %d", commitIndex)
	}

	if err := node.AdvanceCommitIndex(1); err != nil {
		t.Fatal(err)
	}

	start, committed := node.CommittedEntries()
	if start != 0 || len(committed) != 1 {
		t.Fatalf("expected one unapplied committed entry; start=%d entries=%d", start, len(committed))
	}

	if err := node.MarkApplied(1); err != nil {
		t.Fatal(err)
	}

	start, committed = node.CommittedEntries()
	if start != 1 || len(committed) != 0 {
		t.Fatalf("expected no unapplied entries; start=%d entries=%d", start, len(committed))
	}
}

func TestFollowerCannotAppendLeaderEntry(t *testing.T) {
	node := NewNode("node2")

	if _, _, err := node.AppendLeaderEntry("command"); err == nil {
		t.Fatal("expected follower append to fail")
	}
}

func TestCommitIndexCannotExceedLog(t *testing.T) {
	node := NewNode("node1")

	if err := node.AdvanceCommitIndex(1); err == nil {
		t.Fatal("expected commit beyond log length to fail")
	}
}
