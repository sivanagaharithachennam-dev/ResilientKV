package raft

// StartElection moves this node into Candidate state and begins a new term.
// The node votes for itself. Network vote requests will be implemented later.
func (n *Node) StartElection() int {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.CurrentTerm++
	n.State = Candidate
	n.VotedFor = n.ID

	return n.CurrentTerm
}

// BecomeLeader promotes this candidate to Leader if the supplied term
// matches its current term. This method does not itself verify a majority;
// the election coordinator must do that before calling it.
func (n *Node) BecomeLeader(term int) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.State != Candidate || n.CurrentTerm != term {
		return false
	}

	n.State = Leader
	return true
}

// BecomeFollower changes the node to Follower when the observed term
// is at least as new as the node's current term.
func (n *Node) BecomeFollower(term int) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	if term < n.CurrentTerm {
		return false
	}

	if term > n.CurrentTerm {
		n.CurrentTerm = term
		n.VotedFor = ""
	}

	n.State = Follower
	return true
}
