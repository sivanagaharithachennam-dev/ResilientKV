package raft

import "errors"

// AppendLeaderEntry appends a command to the current leader's local log.
// The returned index is a 1-based entry count.
func (n *Node) AppendLeaderEntry(command string) (index int, term int, err error) {
	if n == nil {
		return 0, 0, errors.New("Raft node is nil")
	}
	if command == "" {
		return 0, 0, errors.New("log command cannot be empty")
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if n.State != Leader {
		return 0, n.CurrentTerm, errors.New("only the leader can append commands")
	}

	n.Log = append(n.Log, LogEntry{
		Term:    n.CurrentTerm,
		Command: command,
	})

	return len(n.Log), n.CurrentTerm, nil
}

// LogSnapshot returns a consistent copy of the node's Raft state.
// The returned log can be read without holding the node's mutex.
func (n *Node) LogSnapshot() (
	term int,
	state State,
	entries []LogEntry,
	commitIndex int,
) {
	if n == nil {
		return 0, Follower, nil, 0
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	entries = append([]LogEntry(nil), n.Log...)

	return n.CurrentTerm, n.State, entries, n.CommitIndex
}

// AdvanceCommitIndex advances the commit index without moving it backwards.
// Indexes are represented as the number of committed entries: zero means
// that no entries are committed, one means the first entry is committed.
func (n *Node) AdvanceCommitIndex(index int) error {
	if n == nil {
		return errors.New("Raft node is nil")
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if index < n.CommitIndex {
		return errors.New("commit index cannot move backwards")
	}
	if index > len(n.Log) {
		return errors.New("commit index exceeds log length")
	}

	n.CommitIndex = index
	return nil
}

// CommittedEntries returns a copy of entries not yet marked as applied.
// The first returned value is the number of entries preceding this batch.
func (n *Node) CommittedEntries() (start int, entries []LogEntry) {
	if n == nil {
		return 0, nil
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if n.LastApplied >= n.CommitIndex {
		return n.LastApplied, nil
	}

	entries = append(
		[]LogEntry(nil),
		n.Log[n.LastApplied:n.CommitIndex]...,
	)

	return n.LastApplied, entries
}

// MarkApplied records that all entries up to the supplied entry count
// have been successfully applied to the local state machine.
func (n *Node) MarkApplied(index int) error {
	if n == nil {
		return errors.New("Raft node is nil")
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if index < n.LastApplied {
		return errors.New("last applied index cannot move backwards")
	}
	if index > n.CommitIndex {
		return errors.New("cannot apply entries beyond the commit index")
	}

	n.LastApplied = index
	return nil
}
