package storage

import (
	"fmt"
	"testing"
)

func TestLSMStress(t *testing.T) {
	dataDir := t.TempDir()

	engine, err := New(dataDir)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	defer engine.Close()

	// Use a small MemTable limit so that
	// multiple SSTables are created automatically.
	engine.memTableLimit = 100

	const totalKeys = 1000

	// Insert 1000 keys.
	for i := 0; i < totalKeys; i++ {
		key := fmt.Sprintf("key-%04d", i)
		value := fmt.Sprintf("value-%04d", i)

		if err := engine.Put(key, value); err != nil {
			t.Fatalf("PUT failed for %s: %v", key, err)
		}
	}

	// Flush remaining entries.
	if err := engine.Flush(); err != nil {
		t.Fatalf("final flush failed: %v", err)
	}

	// Verify all keys.
	for i := 0; i < totalKeys; i++ {
		key := fmt.Sprintf("key-%04d", i)
		expected := fmt.Sprintf("value-%04d", i)

		value, exists := engine.Get(key)

		if !exists {
			t.Fatalf("key %s does not exist", key)
		}

		if value != expected {
			t.Fatalf(
				"key %s: expected %s, got %s",
				key,
				expected,
				value,
			)
		}
	}

	// Update some keys.
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key-%04d", i)
		value := fmt.Sprintf("updated-%04d", i)

		if err := engine.Put(key, value); err != nil {
			t.Fatalf("update failed for %s: %v", key, err)
		}
	}

	if err := engine.Flush(); err != nil {
		t.Fatalf("flush after updates failed: %v", err)
	}

	// Delete some keys.
	for i := 200; i < 300; i++ {
		key := fmt.Sprintf("key-%04d", i)

		if err := engine.Delete(key); err != nil {
			t.Fatalf("delete failed for %s: %v", key, err)
		}
	}

	if err := engine.Flush(); err != nil {
		t.Fatalf("flush after deletes failed: %v", err)
	}

	// Verify updated keys.
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key-%04d", i)
		expected := fmt.Sprintf("updated-%04d", i)

		value, exists := engine.Get(key)

		if !exists {
			t.Fatalf("updated key %s does not exist", key)
		}

		if value != expected {
			t.Fatalf(
				"key %s: expected %s, got %s",
				key,
				expected,
				value,
			)
		}
	}

	// Verify deleted keys.
	for i := 200; i < 300; i++ {
		key := fmt.Sprintf("key-%04d", i)

		_, exists := engine.Get(key)

		if exists {
			t.Fatalf(
				"expected deleted key %s to be absent",
				key,
			)
		}
	}

	// Compact all SSTables.
	if err := engine.Compact(); err != nil {
		t.Fatalf("compaction failed: %v", err)
	}

	// Verify data again after compaction.
	for i := 0; i < totalKeys; i++ {
		key := fmt.Sprintf("key-%04d", i)

		value, exists := engine.Get(key)

		if i >= 200 && i < 300 {
			if exists {
				t.Fatalf(
					"deleted key %s exists after compaction",
					key,
				)
			}

			continue
		}

		if !exists {
			t.Fatalf(
				"key %s missing after compaction",
				key,
			)
		}

		if i < 100 {
			expected := fmt.Sprintf("updated-%04d", i)

			if value != expected {
				t.Fatalf(
					"key %s: expected %s, got %s",
					key,
					expected,
					value,
				)
			}
		}
	}

	t.Logf(
		"LSM stress test completed successfully with %d keys",
		totalKeys,
	)
}
