package memtable

import "sync"

// Entry represents one key-value pair.
type Entry struct {
	Key       string
	Value     string
	Tombstone bool
}

// MemTable stores key-value pairs in memory.
type MemTable struct {
	mu   sync.RWMutex
	data map[string]Entry
}

// New creates a new MemTable.
func New() *MemTable {
	return &MemTable{
		data: make(map[string]Entry),
	}
}

// Put inserts or updates a key-value pair.
func (m *MemTable) Put(key string, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = Entry{
		Key:       key,
		Value:     value,
		Tombstone: false,
	}
}

// Delete marks a key as deleted using a tombstone.
func (m *MemTable) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = Entry{
		Key:       key,
		Tombstone: true,
	}
}

// Get retrieves a value using a key.
func (m *MemTable) Get(key string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.data[key]

	if !exists || entry.Tombstone {
		return "", false
	}

	return entry.Value, true
}

// IsDeleted checks whether a key has a tombstone.
func (m *MemTable) IsDeleted(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.data[key]

	return exists && entry.Tombstone
}

// Size returns the number of entries.
func (m *MemTable) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.data)
}

// AllEntries returns all key-value pairs including tombstones.
func (m *MemTable) AllEntries() []Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries := make([]Entry, 0, len(m.data))

	for _, entry := range m.data {
		entries = append(entries, entry)
	}

	return entries
}
