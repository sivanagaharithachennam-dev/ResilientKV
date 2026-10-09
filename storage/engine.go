package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"resilientkv/storage/compaction"
	"resilientkv/storage/memtable"
	"resilientkv/storage/sstable"
	"resilientkv/storage/wal"
)

// Engine is the main storage engine.
type Engine struct {
	mu sync.RWMutex

	memTable *memtable.MemTable
	wal      *wal.WAL
	dataDir  string

	memTableLimit int
	sstableID     int
}

// New creates a new storage engine.
func New(dataDir string) (*Engine, error) {
	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		return nil, err
	}

	walPath := filepath.Join(dataDir, "wal.log")

	w, err := wal.New(walPath)
	if err != nil {
		return nil, err
	}

	engine := &Engine{
		memTable:      memtable.New(),
		wal:           w,
		dataDir:       dataDir,
		memTableLimit: 1000,
		sstableID:     0,
	}

	err = engine.initializeSSTableID()
	if err != nil {
		w.Close()
		return nil, err
	}

	err = engine.recover()
	if err != nil {
		w.Close()
		return nil, err
	}

	return engine, nil
}

// Put stores a key-value pair.
func (e *Engine) Put(key, value string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	operation := fmt.Sprintf("PUT %s %s", key, value)

	if err := e.wal.Append(operation); err != nil {
		return err
	}

	e.memTable.Put(key, value)

	return e.flushIfNeeded()
}

// Delete marks a key as deleted.
func (e *Engine) Delete(key string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	operation := fmt.Sprintf("DELETE %s", key)

	if err := e.wal.Append(operation); err != nil {
		return err
	}

	e.memTable.Delete(key)

	return e.flushIfNeeded()
}

// Get retrieves a value.
func (e *Engine) Get(key string) (string, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	value, exists := e.memTable.Get(key)

	if exists {
		return value, true
	}

	if e.memTable.IsDeleted(key) {
		return "", false
	}

	files, err := os.ReadDir(e.dataDir)
	if err != nil {
		return "", false
	}

	var sstableFiles []string

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".sst" {
			sstableFiles = append(sstableFiles, file.Name())
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(sstableFiles)))

	for _, fileName := range sstableFiles {
		path := filepath.Join(e.dataDir, fileName)

		value, exists, tombstone, err := sstable.Get(path, key)

		if err != nil {
			continue
		}

		if tombstone {
			return "", false
		}

		if exists {
			return value, true
		}
	}

	return "", false
}

// Flush writes the current MemTable to an SSTable
// and replaces it with a new active MemTable.
func (e *Engine) Flush() error {
	entries := e.memTable.AllEntries()

	if len(entries) == 0 {
		return nil
	}

	sstableEntries := make([]sstable.Entry, 0, len(entries))

	for _, entry := range entries {
		sstableEntries = append(sstableEntries, sstable.Entry{
			Key:       entry.Key,
			Value:     entry.Value,
			Tombstone: entry.Tombstone,
		})
	}

	e.memTable = memtable.New()

	e.sstableID++

	sstablePath := filepath.Join(
		e.dataDir,
		fmt.Sprintf("sstable-%06d.sst", e.sstableID),
	)

	return sstable.Write(sstablePath, sstableEntries)
}

// Compact merges all existing SSTables into one SSTable.
func (e *Engine) Compact() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Nothing to compact if there are fewer than two SSTables.
	files, err := os.ReadDir(e.dataDir)
	if err != nil {
		return err
	}

	sstableCount := 0

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".sst" {
			sstableCount++
		}
	}

	if sstableCount < 2 {
		return nil
	}

	// Use the next SSTable ID for the compacted table.
	outputID := e.sstableID + 1

	_, err = compaction.Compact(e.dataDir, outputID)
	if err != nil {
		return err
	}

	e.sstableID = outputID

	return nil
}

// flushIfNeeded checks whether the active MemTable reached its limit.
func (e *Engine) flushIfNeeded() error {
	if e.memTable.Size() < e.memTableLimit {
		return nil
	}

	return e.Flush()
}

// initializeSSTableID finds the highest existing SSTable ID.
func (e *Engine) initializeSSTableID() error {
	files, err := os.ReadDir(e.dataDir)
	if err != nil {
		return err
	}

	maxID := 0

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()

		if !strings.HasPrefix(name, "sstable-") ||
			filepath.Ext(name) != ".sst" {
			continue
		}

		number := strings.TrimSuffix(
			strings.TrimPrefix(name, "sstable-"),
			".sst",
		)

		id, err := strconv.Atoi(number)
		if err != nil {
			continue
		}

		if id > maxID {
			maxID = id
		}
	}

	e.sstableID = maxID

	return nil
}

// recover rebuilds the active MemTable using WAL operations.
func (e *Engine) recover() error {
	operations, err := e.wal.ReadAll()
	if err != nil {
		return err
	}

	for _, operation := range operations {
		parts := strings.SplitN(operation, " ", 3)

		if len(parts) < 2 {
			continue
		}

		command := parts[0]
		key := parts[1]

		switch command {
		case "PUT":
			if len(parts) != 3 {
				continue
			}

			e.memTable.Put(key, parts[2])

		case "DELETE":
			e.memTable.Delete(key)
		}
	}

	return nil
}

// Close closes the storage engine.
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.wal.Close()
}
