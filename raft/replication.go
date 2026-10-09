package raft

import (
	"context"
	"errors"
	"time"

	pb "resilientkv/grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ReplicateLog sends the leader's current log to its followers.
// The returned acknowledgement count includes the leader itself.
// This method does not advance the commit index.
func (n *Node) ReplicateLog(
	peerAddresses []string,
	timeout time.Duration,
) (int, error) {
	if n == nil {
		return 0, errors.New("Raft node is nil")
	}
	if timeout <= 0 {
		return 0, errors.New("replication timeout must be positive")
	}
	if len(peerAddresses) == 0 {
		return 0, errors.New("at least one peer address is required")
	}

	term, state, logEntries, commitIndex := n.LogSnapshot()
	if state != Leader {
		return 0, errors.New("only the leader can replicate its log")
	}

	entries := make([]*pb.ReplicatedLogEntry, 0, len(logEntries))
	for _, entry := range logEntries {
		entries = append(entries, &pb.ReplicatedLogEntry{
			Term:    int32(entry.Term),
			Command: entry.Command,
		})
	}

	// The leader counts as one acknowledgement.
	acknowledgements := 1
	seen := make(map[string]bool)

	for _, address := range peerAddresses {
		if address == "" {
			return acknowledgements, errors.New("peer address cannot be empty")
		}
		if seen[address] {
			return acknowledgements, errors.New("duplicate peer address")
		}
		seen[address] = true
	}

	for _, address := range peerAddresses {
		conn, err := grpc.NewClient(
			address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		response, callErr := pb.NewKeyValueServiceClient(conn).AppendEntries(
			ctx,
			&pb.AppendEntriesRequest{
				Term:         int32(term),
				LeaderId:     n.ID,
				PrevLogIndex: 0,
				PrevLogTerm:  0,
				Entries:      entries,
				LeaderCommit: int32(commitIndex),
			},
		)
		cancel()
		_ = conn.Close()

		if callErr != nil || response == nil {
			continue
		}

		if int(response.GetTerm()) > term {
			n.BecomeFollower(int(response.GetTerm()))
			return acknowledgements, nil
		}

		if response.GetSuccess() {
			acknowledgements++
		}
	}

	return acknowledgements, nil
}
