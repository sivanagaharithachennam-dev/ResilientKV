package raft

import "sync"

// State represents the current role of a Raft node.
type State int

const (
	Follower State = iota
	Candidate
	Leader
)

func (s State) String() string {
	switch s {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	default:
		return "Unknown"
	}
}

// LogEntry represents one command in the Raft replicated log.
type LogEntry struct {
	Term    int
	Command string
}

// Node represents the basic state of a Raft server.
type Node struct {
	mu sync.Mutex

	ID    string
	State State

	CurrentTerm int
	VotedFor    string

	Log         []LogEntry
	CommitIndex int
	LastApplied int
}

// NewNode creates a new Raft node as a follower.
func NewNode(id string) *Node {
	return &Node{
		ID:          id,
		State:       Follower,
		CurrentTerm: 0,
		VotedFor:    "",
		Log:         make([]LogEntry, 0),
		CommitIndex: 0,
		LastApplied: 0,
	}
}
