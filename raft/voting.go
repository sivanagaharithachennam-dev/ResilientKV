package raft

// RequestVote handles a candidate's request for this node's vote.
// The candidate's log must be at least as up-to-date as this node's log.
func (n *Node) RequestVote(
	candidateID string,
	candidateTerm int,
	candidateLastLogIndex int,
	candidateLastLogTerm int,
) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	if candidateID == "" || candidateTerm < n.CurrentTerm {
		return false
	}

	if candidateTerm > n.CurrentTerm {
		n.CurrentTerm = candidateTerm
		n.State = Follower
		n.VotedFor = ""
	}

	lastIndex := len(n.Log)
	lastTerm := 0
	if lastIndex > 0 {
		lastTerm = n.Log[lastIndex-1].Term
	}

	candidateLogIsUpToDate :=
		candidateLastLogTerm > lastTerm ||
			(candidateLastLogTerm == lastTerm &&
				candidateLastLogIndex >= lastIndex)

	if !candidateLogIsUpToDate {
		return false
	}

	if n.VotedFor == "" || n.VotedFor == candidateID {
		n.VotedFor = candidateID
		return true
	}

	return false
}

// UpdateTerm advances the node to a newer term, if applicable.
func (n *Node) UpdateTerm(newTerm int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if newTerm > n.CurrentTerm {
		n.CurrentTerm = newTerm
		n.State = Follower
		n.VotedFor = ""
	}
}
