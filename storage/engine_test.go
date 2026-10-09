package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnginePutGet(t *testing.T) {
	dataDir := t.TempDir()

	engine, err := New(dataDir)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	defer engine.Close()

	err = engine.Put("name", "Yaswanthi")
	if err != nil {
		t.Fatalf("failed to put name: %v", err)
	}

	err = engine.Put("city", "Guntur")
	if err != nil {
		t.Fatalf("failed to put city: %v", err)
	}

	value, exists := engine.Get("name")

	if !exists {
		t.Fatal("expected name to exist")
	}

	if value != "Yaswanthi" {
		t.Fatalf("expected Yaswanthi, got %s", value)
	}

	value, exists = engine.Get("city")

	if !exists {
		t.Fatal("expected city to exist")
	}

	if value != "Guntur" {
		t.Fatalf("expected Guntur, got %s", value)
	}
}

func TestEngineFlush(t *testing.T) {
	dataDir := t.TempDir()

	engine, err := New(dataDir)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	defer engine.Close()

	engine.Put("name", "Yaswanthi")
	engine.Put("city", "Guntur")
	engine.Put("course", "MTech")

	err = engine.Flush()
	if err != nil {
		t.Fatalf("failed to flush: %v", err)
	}

	files, err := os.ReadDir(dataDir)
	if err != nil {
		t.Fatalf("failed to read data directory: %v", err)
	}

	foundSSTable := false

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".sst" {
			foundSSTable = true
			break
		}
	}

	if !foundSSTable {
		t.Fatal("expected SSTable file to be created")
	}
}

func TestEngineWALRecovery(t *testing.T) {
	dataDir := t.TempDir()

	// Start the first engine.
	engine, err := New(dataDir)
	if err != nil {
		t.Fatalf("failed to create first engine: %v", err)
	}

	err = engine.Put("name", "Yaswanthi")
	if err != nil {
		t.Fatalf("failed to put name: %v", err)
	}

	err = engine.Put("city", "Guntur")
	if err != nil {
		t.Fatalf("failed to put city: %v", err)
	}

	// Close the first engine.
	err = engine.Close()
	if err != nil {
		t.Fatalf("failed to close first engine: %v", err)
	}

	// Start a new engine using the same directory.
	recoveredEngine, err := New(dataDir)
	if err != nil {
		t.Fatalf("failed to create recovered engine: %v", err)
	}

	defer recoveredEngine.Close()

	// Check recovered data.
	value, exists := recoveredEngine.Get("name")

	if !exists {
		t.Fatal("expected name to be recovered")
	}

	if value != "Yaswanthi" {
		t.Fatalf("expected Yaswanthi, got %s", value)
	}

	value, exists = recoveredEngine.Get("city")

	if !exists {
		t.Fatal("expected city to be recovered")
	}

	if value != "Guntur" {
		t.Fatalf("expected Guntur, got %s", value)
	}
}

func TestEngineDelete(t *testing.T) {
	dataDir := t.TempDir()

	engine, err := New(dataDir)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	defer engine.Close()

	err = engine.Put("name", "Yaswanthi")
	if err != nil {
		t.Fatalf("failed to put name: %v", err)
	}

	err = engine.Delete("name")
	if err != nil {
		t.Fatalf("failed to delete name: %v", err)
	}

	_, exists := engine.Get("name")

	if exists {
		t.Fatal("expected name to be deleted")
	}
}

func TestEngineDeleteRecovery(t *testing.T) {
	dataDir := t.TempDir()

	engine, err := New(dataDir)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	err = engine.Put("name", "Yaswanthi")
	if err != nil {
		t.Fatalf("failed to put name: %v", err)
	}

	err = engine.Delete("name")
	if err != nil {
		t.Fatalf("failed to delete name: %v", err)
	}

	err = engine.Close()
	if err != nil {
		t.Fatalf("failed to close engine: %v", err)
	}

	// Start the engine again.
	recoveredEngine, err := New(dataDir)
	if err != nil {
		t.Fatalf("failed to recover engine: %v", err)
	}

	defer recoveredEngine.Close()

	_, exists := recoveredEngine.Get("name")

	if exists {
		t.Fatal("expected deleted key to remain deleted after recovery")
	}
}

func TestEngineAutomaticFlush(t *testing.T) {
	dataDir := t.TempDir()

	engine, err := New(dataDir)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	defer engine.Close()

	// Use a very small limit for testing.
	engine.memTableLimit = 2

	err = engine.Put("a", "1")
	if err != nil {
		t.Fatalf("failed to put a: %v", err)
	}

	err = engine.Put("b", "2")
	if err != nil {
		t.Fatalf("failed to put b: %v", err)
	}

	files, err := os.ReadDir(dataDir)
	if err != nil {
		t.Fatalf("failed to read data directory: %v", err)
	}

	foundSSTable := false

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".sst" {
			foundSSTable = true
			break
		}
	}

	if !foundSSTable {
		t.Fatal("expected automatic flush to create an SSTable")
	}
}

func TestEnginePersistentTombstoneRecovery(t *testing.T) {
	dataDir := t.TempDir()

	// Create the first engine.
	engine, err := New(dataDir)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Store the key.
	err = engine.Put("user", "Yaswanthi")
	if err != nil {
		t.Fatalf("failed to put user: %v", err)
	}

	// Flush the value to an SSTable.
	err = engine.Flush()
	if err != nil {
		t.Fatalf("failed to flush value: %v", err)
	}

	// Delete the key.
	err = engine.Delete("user")
	if err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}

	// Flush the tombstone to an SSTable.
	err = engine.Flush()
	if err != nil {
		t.Fatalf("failed to flush tombstone: %v", err)
	}

	// Close the first engine.
	err = engine.Close()
	if err != nil {
		t.Fatalf("failed to close engine: %v", err)
	}

	// Create a new engine using the same directory.
	engine2, err := New(dataDir)
	if err != nil {
		t.Fatalf("failed to recover engine: %v", err)
	}

	defer engine2.Close()

	// The deleted key must remain deleted after restart.
	value, exists := engine2.Get("user")

	if exists {
		t.Fatalf(
			"expected user to remain deleted after recovery, got %s",
			value,
		)
	}
}

func TestEnginePersistentSSTableID(t *testing.T) {
	dataDir := t.TempDir()

	// First engine.
	engine, err := New(dataDir)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	err = engine.Put("key1", "value1")
	if err != nil {
		t.Fatalf("failed to put key1: %v", err)
	}

	err = engine.Flush()
	if err != nil {
		t.Fatalf("failed to flush first SSTable: %v", err)
	}

	if engine.sstableID != 1 {
		t.Fatalf("expected SSTable ID 1, got %d", engine.sstableID)
	}

	err = engine.Close()
	if err != nil {
		t.Fatalf("failed to close first engine: %v", err)
	}

	// Restart using the same data directory.
	engine2, err := New(dataDir)
	if err != nil {
		t.Fatalf("failed to restart engine: %v", err)
	}

	defer engine2.Close()

	if engine2.sstableID != 1 {
		t.Fatalf(
			"expected recovered SSTable ID 1, got %d",
			engine2.sstableID,
		)
	}

	// Create another SSTable.
	err = engine2.Put("key2", "value2")
	if err != nil {
		t.Fatalf("failed to put key2: %v", err)
	}

	err = engine2.Flush()
	if err != nil {
		t.Fatalf("failed to flush second SSTable: %v", err)
	}

	if engine2.sstableID != 2 {
		t.Fatalf(
			"expected new SSTable ID 2, got %d",
			engine2.sstableID,
		)
	}

	// Verify both SSTable files exist.
	if _, err := os.Stat(filepath.Join(dataDir, "sstable-000001.sst")); err != nil {
		t.Fatalf("first SSTable not found: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dataDir, "sstable-000002.sst")); err != nil {
		t.Fatalf("second SSTable not found: %v", err)
	}
}

func TestEngineCompaction(t *testing.T) {
	dataDir := t.TempDir()

	engine, err := New(dataDir)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	defer engine.Close()

	// Create first SSTable.
	err = engine.Put("name", "Old")
	if err != nil {
		t.Fatalf("failed to put Old: %v", err)
	}

	err = engine.Flush()
	if err != nil {
		t.Fatalf("failed to flush first SSTable: %v", err)
	}

	// Create second SSTable with newer value.
	err = engine.Put("name", "New")
	if err != nil {
		t.Fatalf("failed to put New: %v", err)
	}

	err = engine.Flush()
	if err != nil {
		t.Fatalf("failed to flush second SSTable: %v", err)
	}

	// Compact both SSTables.
	err = engine.Compact()
	if err != nil {
		t.Fatalf("compaction failed: %v", err)
	}

	// Newest value must remain.
	value, exists := engine.Get("name")

	if !exists {
		t.Fatal("expected name to exist after compaction")
	}

	if value != "New" {
		t.Fatalf(
			"expected New after compaction, got %s",
			value,
		)
	}

	// Only one SSTable should remain.
	files, err := os.ReadDir(dataDir)
	if err != nil {
		t.Fatalf("failed to read data directory: %v", err)
	}

	sstableCount := 0

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".sst" {
			sstableCount++
		}
	}

	if sstableCount != 1 {
		t.Fatalf(
			"expected 1 SSTable after compaction, got %d",
			sstableCount,
		)
	}
}

func TestEngineCompactionTombstoneRecovery(t *testing.T) {
	dataDir := t.TempDir()

	// First engine.
	engine, err := New(dataDir)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Store the value.
	err = engine.Put("user", "Yaswanthi")
	if err != nil {
		t.Fatalf("failed to put user: %v", err)
	}

	err = engine.Flush()
	if err != nil {
		t.Fatalf("failed to flush value: %v", err)
	}

	// Delete the value.
	err = engine.Delete("user")
	if err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}

	err = engine.Flush()
	if err != nil {
		t.Fatalf("failed to flush tombstone: %v", err)
	}

	// Compact the value and tombstone.
	err = engine.Compact()
	if err != nil {
		t.Fatalf("compaction failed: %v", err)
	}

	// The key must remain deleted.
	_, exists := engine.Get("user")

	if exists {
		t.Fatal("expected user to remain deleted after compaction")
	}

	// Close the first engine.
	err = engine.Close()
	if err != nil {
		t.Fatalf("failed to close engine: %v", err)
	}

	// Restart using the same data directory.
	engine2, err := New(dataDir)
	if err != nil {
		t.Fatalf("failed to restart engine: %v", err)
	}

	defer engine2.Close()

	// The deleted key must still be deleted after restart.
	value, exists := engine2.Get("user")

	if exists {
		t.Fatalf(
			"expected user to remain deleted after restart, got %s",
			value,
		)
	}
}
