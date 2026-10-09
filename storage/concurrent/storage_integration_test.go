package concurrent

import (
	"fmt"
	"sync"
	"testing"

	"resilientkv/storage"
)

func TestConcurrentStorageIntegration(t *testing.T) {
	dataDir := t.TempDir()

	engine, err := storage.New(dataDir)
	if err != nil {
		t.Fatalf("failed to create storage engine: %v", err)
	}

	defer engine.Close()

	const goroutines = 20
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup

	for g := 0; g < goroutines; g++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			for i := 0; i < operationsPerGoroutine; i++ {
				key := fmt.Sprintf("key-%d-%d", id, i)

				value := fmt.Sprintf("value-%d-%d", id, i)

				if err := engine.Put(key, value); err != nil {
					t.Errorf("Put failed for %s: %v", key, err)
					return
				}

				got, exists := engine.Get(key)

				if !exists {
					t.Errorf("key %s was not found", key)
					return
				}

				if got != value {
					t.Errorf(
						"wrong value for %s: expected %s, got %s",
						key,
						value,
						got,
					)
					return
				}
			}
		}(g)
	}

	wg.Wait()

	t.Logf(
		"completed %d goroutines × %d operations",
		goroutines,
		operationsPerGoroutine,
	)
}
