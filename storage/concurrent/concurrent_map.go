package concurrent

import (
	"hash/fnv"
	"sync"
)

// shard stores a portion of the key-value data.
type shard struct {
	mu   sync.RWMutex
	data map[string]string
}

// ConcurrentMap is a sharded concurrent key-value map.
type ConcurrentMap struct {
	shards []shard
}

// NewConcurrentMap creates a new concurrent map.
func NewConcurrentMap(shardCount int) *ConcurrentMap {
	if shardCount <= 0 {
		shardCount = 16
	}

	cm := &ConcurrentMap{
		shards: make([]shard, shardCount),
	}

	for i := range cm.shards {
		cm.shards[i].data = make(map[string]string)
	}

	return cm
}

// getShard returns the shard responsible for a key.
func (cm *ConcurrentMap) getShard(key string) *shard {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(key))

	index := int(hash.Sum32()) % len(cm.shards)

	return &cm.shards[index]
}

// Put inserts or updates a key-value pair.
func (cm *ConcurrentMap) Put(key string, value string) {
	s := cm.getShard(key)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = value
}

// Get retrieves a value.
func (cm *ConcurrentMap) Get(key string) (string, bool) {
	s := cm.getShard(key)

	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]

	return value, exists
}

// Delete removes a key.
func (cm *ConcurrentMap) Delete(key string) {
	s := cm.getShard(key)

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
}

// Size returns the total number of entries.
func (cm *ConcurrentMap) Size() int {
	total := 0

	for i := range cm.shards {
		s := &cm.shards[i]

		s.mu.RLock()
		total += len(s.data)
		s.mu.RUnlock()
	}

	return total
}
