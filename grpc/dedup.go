package grpc

import "sync"

type Deduplicator struct {
	mu        sync.Mutex
	processed map[string]bool
}

func NewDeduplicator() *Deduplicator {
	return &Deduplicator{
		processed: make(map[string]bool),
	}
}

func (d *Deduplicator) IsProcessed(requestID string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.processed[requestID]
}

func (d *Deduplicator) MarkProcessed(requestID string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.processed[requestID] = true
}
