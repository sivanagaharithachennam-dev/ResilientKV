package raft

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "resilientkv/grpc/proto"
)

// ConductNetworkElection requests votes from other Raft servers.
// The node becomes leader only if it receives a cluster majority.
func (n *Node) ConductNetworkElection(
	peerAddresses []string,
	timeout time.Duration,
) (bool, error) {
	if n == nil {
		return false, errors.New("Raft node is nil")
	}
	if timeout <= 0 {
		return false, errors.New("timeout must be greater than zero")
	}

	// Validate the peer addresses before starting an election.
	seen := make(map[string]bool)
	for _, address := range peerAddresses {
		if address == "" {
			return false, errors.New("peer address cannot be empty")
		}
		if seen[address] {
			return false, fmt.Errorf("duplicate peer address: %s", address)
		}
		seen[address] = true
	}

	// Start a new term and vote for this candidate.
	term := n.StartElection()

	// Read the candidate's last log position safely.
	n.mu.Lock()
	lastIndex := len(n.Log)
	lastTerm := 0
	if lastIndex > 0 {
		lastTerm = n.Log[lastIndex-1].Term
	}
	candidateID := n.ID
	n.mu.Unlock()

	if candidateID == "" {
		return false, errors.New("candidate ID cannot be empty")
	}

	votes := 1 // The candidate votes for itself.
	majority := (len(peerAddresses)+1)/2 + 1

	for _, address := range peerAddresses {
		conn, err := grpc.NewClient(
			address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			// An unreachable peer does not contribute a vote.
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		client := pb.NewKeyValueServiceClient(conn)

		response, callErr := client.RequestVote(ctx, &pb.RequestVoteRequest{
			CandidateId:  candidateID,
			Term:         int32(term),
			LastLogIndex: int32(lastIndex),
			LastLogTerm:  int32(lastTerm),
		})

		cancel()
		_ = conn.Close()

		if callErr != nil {
			// A timeout or connection error counts as no vote.
			continue
		}

		// If a server reports a newer term, step down immediately.
		if int(response.GetTerm()) > term {
			n.BecomeFollower(int(response.GetTerm()))
			return false, nil
		}

		if response.GetVoteGranted() {
			votes++
		}
	}

	// Check whether this node still belongs to the same election.
	n.mu.Lock()
	stillCandidate := n.State == Candidate && n.CurrentTerm == term
	n.mu.Unlock()

	if !stillCandidate || votes < majority {
		return false, nil
	}

	return n.BecomeLeader(term), nil
}
