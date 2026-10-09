package raft

import (
	"context"
	"errors"
	"sync"
	"time"

	pb "resilientkv/grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// HeartbeatConfig controls the leader's heartbeat loop.
type HeartbeatConfig struct {
	Interval time.Duration
	Timeout  time.Duration
}

// StartHeartbeats periodically sends empty AppendEntries requests to followers.
// Call cancel() on the returned context to stop the heartbeat loop.
func (n *Node) StartHeartbeats(
	parent context.Context,
	peerAddresses []string,
	config HeartbeatConfig,
) (context.CancelFunc, error) {
	if n == nil {
		return nil, errors.New("Raft node is nil")
	}
	if parent == nil {
		return nil, errors.New("parent context cannot be nil")
	}
	if config.Interval <= 0 || config.Timeout <= 0 {
		return nil, errors.New("heartbeat interval and timeout must be positive")
	}

	peers := make([]string, 0, len(peerAddresses))
	seen := make(map[string]bool)
	for _, address := range peerAddresses {
		if address == "" {
			return nil, errors.New("peer address cannot be empty")
		}
		if seen[address] {
			return nil, errors.New("duplicate peer address")
		}
		seen[address] = true
		peers = append(peers, address)
	}
	if len(peers) == 0 {
		return nil, errors.New("at least one peer address is required")
	}

	ctx, cancel := context.WithCancel(parent)

	var once sync.Once
	stop := func() {
		once.Do(cancel)
	}

	go func() {
		ticker := time.NewTicker(config.Interval)
		defer ticker.Stop()

		// Send the first heartbeat immediately.
		n.sendHeartbeats(ctx, peers, config.Timeout)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				n.sendHeartbeats(ctx, peers, config.Timeout)
			}
		}
	}()

	return stop, nil
}

func (n *Node) sendHeartbeats(
	ctx context.Context,
	peerAddresses []string,
	timeout time.Duration,
) {
	if n.GetState() != Leader {
		return
	}

	term := n.GetCurrentTerm()

	n.mu.Lock()
	lastIndex := len(n.Log)
	lastTerm := 0
	if lastIndex > 0 {
		lastTerm = n.Log[lastIndex-1].Term
	}
	commitIndex := n.CommitIndex
	n.mu.Unlock()

	for _, address := range peerAddresses {
		if ctx.Err() != nil {
			return
		}

		conn, err := grpc.NewClient(
			address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			continue
		}

		callCtx, cancel := context.WithTimeout(ctx, timeout)
		response, callErr := pb.NewKeyValueServiceClient(conn).AppendEntries(
			callCtx,
			&pb.AppendEntriesRequest{
				Term:         int32(term),
				LeaderId:     n.ID,
				PrevLogIndex: int32(lastIndex),
				PrevLogTerm:  int32(lastTerm),
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
			return
		}
	}
}
