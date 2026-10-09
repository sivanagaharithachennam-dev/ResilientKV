package raft

// ConductElection simulates a Raft election among nodes in one process.
// The candidate becomes leader only if it receives a majority of votes.
// Network communication will be added in a later step.
func ConductElection(nodes []*Node, candidate *Node) bool {
	if candidate == nil || len(nodes) == 0 {
		return false
	}

	// Validate the cluster and make sure the candidate is a member.
	seenNodes := make(map[*Node]bool)
	seenIDs := make(map[string]bool)
	candidateFound := false

	for _, node := range nodes {
		if node == nil || node.ID == "" ||
			seenNodes[node] || seenIDs[node.ID] {
			return false
		}

		seenNodes[node] = true
		seenIDs[node.ID] = true

		if node == candidate {
			candidateFound = true
		}
	}

	if !candidateFound {
		return false
	}

	// Start a new election term and vote for itself.
	term := candidate.StartElection()

	candidate.mu.Lock()
	lastIndex := len(candidate.Log)
	lastTerm := 0
	if lastIndex > 0 {
		lastTerm = candidate.Log[lastIndex-1].Term
	}
	candidate.mu.Unlock()

	votes := 1 // The candidate's own vote.
	majority := len(nodes)/2 + 1

	for _, peer := range nodes {
		if peer == candidate {
			continue
		}

		granted := peer.RequestVote(
			candidate.ID,
			term,
			lastIndex,
			lastTerm,
		)

		// A higher term means this candidate must step down.
		peer.mu.Lock()
		peerTerm := peer.CurrentTerm
		peer.mu.Unlock()

		if peerTerm > term {
			candidate.BecomeFollower(peerTerm)
			return false
		}

		if granted {
			votes++
		}
	}

	// Never declare a leader without a majority.
	if votes < majority {
		return false
	}

	return candidate.BecomeLeader(term)
}
