package grpc

import (
	"errors"
	"sync"

	"resilientkv/raft"
	"resilientkv/storage"
)

// commandApplier serializes application of committed Raft entries.
type commandApplier struct {
	mu     sync.Mutex
	node   *raft.Node
	engine *storage.Engine
}

// applyCommitted applies committed commands in log order.
// The caller must provide a configured node and storage engine.
func (a *commandApplier) applyCommitted() error {
	if a == nil || a.node == nil || a.engine == nil {
		return errors.New("command applier is not configured")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	start, entries := a.node.CommittedEntries()
	for i, entry := range entries {
		command, err := raft.DecodeCommand(entry.Command)
		if err != nil {
			return err
		}

		switch command.Op {
		case "put":
			err = a.engine.Put(command.Key, command.Value)
		case "delete":
			err = a.engine.Delete(command.Key)
		default:
			err = errors.New("unsupported committed command")
		}
		if err != nil {
			return err
		}

		if err := a.node.MarkApplied(start + i + 1); err != nil {
			return err
		}
	}

	return nil
}
