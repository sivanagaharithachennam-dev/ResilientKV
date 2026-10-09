package raft

import "errors"

// Majority returns the number of votes or acknowledgements required
// for a strict majority of a cluster.
func Majority(clusterSize int) (int, error) {
	if clusterSize <= 0 {
		return 0, errors.New("cluster size must be positive")
	}
	return clusterSize/2 + 1, nil
}

// HasMajority reports whether acknowledgements form a strict majority.
func HasMajority(acknowledgements, clusterSize int) (bool, error) {
	required, err := Majority(clusterSize)
	if err != nil {
		return false, err
	}
	if acknowledgements < 0 || acknowledgements > clusterSize {
		return false, errors.New("acknowledgements must be between zero and cluster size")
	}
	return acknowledgements >= required, nil
}

// AdvanceCommitWithMajority advances the leader's commit index when the
// supplied acknowledgement count is a majority and the target entry belongs
// to the leader's current term.
//
// The acknowledgement count must represent distinct members of the same
// configured cluster, including the leader itself when applicable.
// This method does not collect acknowledgements or apply entries to storage.
func (n *Node) AdvanceCommitWithMajority(index, acknowledgements, clusterSize int) (bool, error) {
	if n == nil {
		return false, errors.New("Raft node is nil")
	}

	required, err := Majority(clusterSize)
	if err != nil {
		return false, err
	}
	if acknowledgements < 0 || acknowledgements > clusterSize {
		return false, errors.New("invalid acknowledgement count")
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if n.State != Leader {
		return false, errors.New("only the leader can advance commitment")
	}
	if index <= n.CommitIndex {
		return false, errors.New("target index must exceed the current commit index")
	}
	if index > len(n.Log) {
		return false, errors.New("target index exceeds log length")
	}
	if acknowledgements < required {
		return false, nil
	}

	// Raft leaders directly advance commitment using a majority only for
	// entries from their current term.
	if n.Log[index-1].Term != n.CurrentTerm {
		return false, nil
	}

	n.CommitIndex = index
	return true, nil
}
