package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	pb "resilientkv/grpc/proto"
	"resilientkv/raft"
	"resilientkv/storage"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedKeyValueServiceServer

	engine       *storage.Engine
	deduplicator *Deduplicator
	raftNode     *raft.Node
	applier      *commandApplier

	// Peers contains the configured addresses of other cluster members.
	peerMu sync.RWMutex
	peers  []string

	// Serialize leader writes so their log/commit/apply steps do not overlap.
	writeMu sync.Mutex

	heartbeatMu     sync.Mutex
	heartbeatCancel context.CancelFunc
}

func NewServer(engine *storage.Engine) *Server {
	return newServer(engine, nil)
}

func NewServerWithRaft(engine *storage.Engine, node *raft.Node) *Server {
	return newServer(engine, node)
}

func newServer(engine *storage.Engine, node *raft.Node) *Server {
	server := &Server{
		engine:       engine,
		deduplicator: NewDeduplicator(),
		raftNode:     node,
	}

	if node != nil && engine != nil {
		server.applier = &commandApplier{
			node:   node,
			engine: engine,
		}
	}

	return server
}

// SetPeers configures the addresses of the other Raft cluster members.
func (s *Server) SetPeers(addresses []string) {
	peers := append([]string(nil), addresses...)

	s.peerMu.Lock()
	s.peers = peers
	s.peerMu.Unlock()
}

// peerAddresses returns a copy of the configured peer addresses.
func (s *Server) peerAddresses() []string {
	s.peerMu.RLock()
	defer s.peerMu.RUnlock()

	return append([]string(nil), s.peers...)
}

// executeRaftWrite serializes leader writes and commits them only after
// receiving acknowledgements from a majority of the configured cluster.
func (s *Server) executeRaftWrite(command raft.Command, requestID string) error {
	if s.raftNode == nil || s.engine == nil || s.applier == nil {
		return fmt.Errorf("Raft node and storage engine must be configured")
	}

	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	if requestID != "" && s.deduplicator.IsProcessed(requestID) {
		return nil
	}

	if s.raftNode.GetState() != raft.Leader {
		return fmt.Errorf("not the Raft leader; current state is %s", s.raftNode.GetState())
	}

	encoded, err := raft.EncodeCommand(command)
	if err != nil {
		return fmt.Errorf("encode command: %w", err)
	}

	index, _, err := s.raftNode.AppendLeaderEntry(encoded)
	if err != nil {
		return fmt.Errorf("append leader entry: %w", err)
	}

	peers := s.peerAddresses()
	acknowledgements := 1 // The leader counts as one member.

	if len(peers) > 0 {
		acknowledgements, err = s.raftNode.ReplicateLog(peers, 2*time.Second)
		if err != nil {
			return fmt.Errorf("replicate log: %w", err)
		}
	}

	clusterSize := len(peers) + 1
	committed, err := s.raftNode.AdvanceCommitWithMajority(
		index, acknowledgements, clusterSize,
	)
	if err != nil {
		return fmt.Errorf("commit log entry: %w", err)
	}
	if !committed {
		required, _ := raft.Majority(clusterSize)
		return fmt.Errorf(
			"write not committed: received %d acknowledgements; %d required",
			acknowledgements, required,
		)
	}

	if err := s.applier.applyCommitted(); err != nil {
		return fmt.Errorf("apply committed command: %w", err)
	}

	if requestID != "" {
		s.deduplicator.MarkProcessed(requestID)
	}

	// Notify followers of the updated commit index. The write is already
	// committed, so a notification failure does not undo the successful write.
	if len(peers) > 0 {
		if _, notifyErr := s.raftNode.ReplicateLog(peers, 2*time.Second); notifyErr != nil {
			// A later heartbeat can propagate the updated commit index.
		}
	}

	return nil
}

func (s *Server) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	if req == nil || req.GetKey() == "" {
		return &pb.PutResponse{Success: false, Error: "key cannot be empty"}, nil
	}

	if s.raftNode != nil {
		err := s.executeRaftWrite(raft.Command{
			Op:    "put",
			Key:   req.GetKey(),
			Value: req.GetValue(),
		}, req.GetRequestId())
		if err != nil {
			return &pb.PutResponse{Success: false, Error: err.Error()}, nil
		}
		return &pb.PutResponse{Success: true}, nil
	}

	if req.GetRequestId() != "" && s.deduplicator.IsProcessed(req.GetRequestId()) {
		return &pb.PutResponse{Success: true}, nil
	}
	if s.engine == nil {
		return &pb.PutResponse{Success: false, Error: "storage engine is not configured"}, nil
	}
	if err := s.engine.Put(req.GetKey(), req.GetValue()); err != nil {
		return &pb.PutResponse{Success: false, Error: err.Error()}, nil
	}
	if req.GetRequestId() != "" {
		s.deduplicator.MarkProcessed(req.GetRequestId())
	}
	return &pb.PutResponse{Success: true}, nil
}

func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	value, exists := s.engine.Get(req.GetKey())
	return &pb.GetResponse{
		Value: value,
		Found: exists,
	}, nil
}

func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	if req == nil || req.GetKey() == "" {
		return &pb.DeleteResponse{Success: false, Error: "key cannot be empty"}, nil
	}

	if s.raftNode != nil {
		err := s.executeRaftWrite(raft.Command{
			Op:  "delete",
			Key: req.GetKey(),
		}, req.GetRequestId())
		if err != nil {
			return &pb.DeleteResponse{Success: false, Error: err.Error()}, nil
		}
		return &pb.DeleteResponse{Success: true}, nil
	}

	if req.GetRequestId() != "" && s.deduplicator.IsProcessed(req.GetRequestId()) {
		return &pb.DeleteResponse{Success: true}, nil
	}
	if s.engine == nil {
		return &pb.DeleteResponse{Success: false, Error: "storage engine is not configured"}, nil
	}
	if err := s.engine.Delete(req.GetKey()); err != nil {
		return &pb.DeleteResponse{Success: false, Error: err.Error()}, nil
	}
	if req.GetRequestId() != "" {
		s.deduplicator.MarkProcessed(req.GetRequestId())
	}
	return &pb.DeleteResponse{Success: true}, nil
}

func (s *Server) RequestVote(
	ctx context.Context,
	req *pb.RequestVoteRequest,
) (*pb.RequestVoteResponse, error) {
	if req == nil || req.GetCandidateId() == "" || req.GetTerm() < 0 ||
		req.GetLastLogIndex() < 0 || req.GetLastLogTerm() < 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid vote request")
	}

	if s.raftNode == nil {
		return nil, status.Error(codes.FailedPrecondition, "Raft node is not configured")
	}

	granted := s.raftNode.RequestVote(
		req.GetCandidateId(),
		int(req.GetTerm()),
		int(req.GetLastLogIndex()),
		int(req.GetLastLogTerm()),
	)

	return &pb.RequestVoteResponse{
		Term:        int32(s.raftNode.GetCurrentTerm()),
		VoteGranted: granted,
	}, nil
}

// TriggerElection starts a network election and starts heartbeats if elected.
func (s *Server) TriggerElection(
	ctx context.Context,
	req *pb.TriggerElectionRequest,
) (*pb.TriggerElectionResponse, error) {
	if req == nil || len(req.GetPeerAddresses()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one peer address is required")
	}

	timeoutMillis := req.GetTimeoutMillis()
	if timeoutMillis <= 0 || timeoutMillis > 30000 {
		return nil, status.Error(codes.InvalidArgument, "timeout_millis must be between 1 and 30000")
	}

	if s.raftNode == nil {
		return nil, status.Error(codes.FailedPrecondition, "Raft node is not configured")
	}

	elected, err := s.raftNode.ConductNetworkElection(
		req.GetPeerAddresses(),
		time.Duration(timeoutMillis)*time.Millisecond,
	)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "election could not start: %v", err)
	}

	if elected {
		s.heartbeatMu.Lock()

		if s.heartbeatCancel != nil {
			s.heartbeatCancel()
			s.heartbeatCancel = nil
		}

		heartbeatCtx, cancel := context.WithCancel(context.Background())
		stop, startErr := s.raftNode.StartHeartbeats(
			heartbeatCtx,
			req.GetPeerAddresses(),
			raft.HeartbeatConfig{
				Interval: 500 * time.Millisecond,
				Timeout:  300 * time.Millisecond,
			},
		)
		if startErr != nil {
			cancel()
			s.heartbeatMu.Unlock()
			return nil, status.Errorf(codes.Internal, "start heartbeats: %v", startErr)
		}

		s.heartbeatCancel = func() {
			cancel()
			stop()
		}
		s.heartbeatMu.Unlock()
	}

	return &pb.TriggerElectionResponse{
		Elected: elected,
		Term:    int32(s.raftNode.GetCurrentTerm()),
		State:   s.raftNode.GetState().String(),
	}, nil
}

// AppendEntries handles a Raft heartbeat or log replication request.
func (s *Server) AppendEntries(
	ctx context.Context,
	req *pb.AppendEntriesRequest,
) (*pb.AppendEntriesResponse, error) {
	if req == nil || req.GetLeaderId() == "" ||
		req.GetTerm() < 0 ||
		req.GetPrevLogIndex() < 0 ||
		req.GetPrevLogTerm() < 0 ||
		req.GetLeaderCommit() < 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid AppendEntries request")
	}

	if s.raftNode == nil {
		return nil, status.Error(codes.FailedPrecondition, "Raft node is not configured")
	}

	entries := make([]raft.LogEntry, 0, len(req.GetEntries()))
	for _, entry := range req.GetEntries() {
		if entry == nil || entry.GetTerm() < 0 || entry.GetTerm() > req.GetTerm() {
			return nil, status.Error(codes.InvalidArgument, "invalid replicated log entry")
		}

		entries = append(entries, raft.LogEntry{
			Term:    int(entry.GetTerm()),
			Command: entry.GetCommand(),
		})
	}

	result, err := s.raftNode.HandleAppendEntries(raft.AppendEntriesRequest{
		Term:         int(req.GetTerm()),
		LeaderID:     req.GetLeaderId(),
		PrevLogIndex: int(req.GetPrevLogIndex()),
		PrevLogTerm:  int(req.GetPrevLogTerm()),
		Entries:      entries,
		LeaderCommit: int(req.GetLeaderCommit()),
	})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "AppendEntries failed: %v", err)
	}

	if result.Success && s.applier != nil {
		if err := s.applier.applyCommitted(); err != nil {
			return nil, status.Errorf(codes.Internal, "apply committed entries: %v", err)
		}
	}

	return &pb.AppendEntriesResponse{
		Term:    int32(result.Term),
		Success: result.Success,
	}, nil
}
