package raft

import "errors"

// AppendEntriesRequest contains the information sent by a Raft leader.
type AppendEntriesRequest struct {
	Term         int
	LeaderID     string
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

// AppendEntriesResponse reports whether the follower accepted the request.
type AppendEntriesResponse struct {
	Term    int
	Success bool
}

// HandleAppendEntries processes a leader heartbeat or log replication request.
// PrevLogIndex is the number of entries preceding the incoming entries.
func (n *Node) HandleAppendEntries(req AppendEntriesRequest) (AppendEntriesResponse, error) {
	if n == nil {
		return AppendEntriesResponse{}, errors.New("Raft node is nil")
	}
	if req.LeaderID == "" {
		return AppendEntriesResponse{}, errors.New("leader ID cannot be empty")
	}
	if req.Term < 0 || req.PrevLogIndex < 0 ||
		req.PrevLogTerm < 0 || req.LeaderCommit < 0 {
		return AppendEntriesResponse{}, errors.New("Raft request contains a negative value")
	}

	for _, entry := range req.Entries {
		if entry.Term < 0 || entry.Term > req.Term {
			return AppendEntriesResponse{}, errors.New("log entry term is invalid")
		}
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	// Reject requests from an outdated leader.
	if req.Term < n.CurrentTerm {
		return AppendEntriesResponse{Term: n.CurrentTerm, Success: false}, nil
	}

	// A current-term or newer leader makes this node a follower.
	if req.Term > n.CurrentTerm {
		n.CurrentTerm = req.Term
		n.VotedFor = ""
	}
	n.State = Follower

	// Verify that the follower contains the requested log prefix.
	if req.PrevLogIndex > len(n.Log) {
		return AppendEntriesResponse{Term: n.CurrentTerm, Success: false}, nil
	}
	if req.PrevLogIndex > 0 {
		localPrev := n.Log[req.PrevLogIndex-1]
		if localPrev.Term != req.PrevLogTerm {
			return AppendEntriesResponse{Term: n.CurrentTerm, Success: false}, nil
		}
	}

	// Reconcile incoming entries, replacing only a conflicting uncommitted
	// suffix. CommitIndex is a count of committed entries.
	for i, incoming := range req.Entries {
		index := req.PrevLogIndex + i

		if index >= len(n.Log) {
			n.Log = append(n.Log, req.Entries[i:]...)
			break
		}

		local := n.Log[index]

		if local.Term == incoming.Term {
			// In a valid Raft log, equal index and term imply the same entry.
			// Reject a different command rather than silently keeping it.
			if local.Command != incoming.Command {
				return AppendEntriesResponse{Term: n.CurrentTerm, Success: false}, nil
			}
			continue
		}

		// Never overwrite an entry that has already been committed.
		if index < n.CommitIndex {
			return AppendEntriesResponse{Term: n.CurrentTerm, Success: false}, nil
		}

		n.Log = append(n.Log[:index], req.Entries[i:]...)
		break
	}

	// Advance commitment only as far as the follower's log permits.
	if req.LeaderCommit > n.CommitIndex {
		n.CommitIndex = req.LeaderCommit
		if n.CommitIndex > len(n.Log) {
			n.CommitIndex = len(n.Log)
		}
	}

	return AppendEntriesResponse{Term: n.CurrentTerm, Success: true}, nil
}
