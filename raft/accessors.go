package raft

// GetCurrentTerm returns the current term safely.
func (n *Node) GetCurrentTerm() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.CurrentTerm
}

// GetState returns the current node state safely.
func (n *Node) GetState() State {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.State
}
