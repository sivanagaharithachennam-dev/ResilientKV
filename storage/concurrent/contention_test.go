package concurrent

import (
	"fmt"
	"sync"
	"testing"
)

func TestConcurrentMapHighContention(t *testing.T) {
	m := NewConcurrentMap(32)

	const goroutines = 100
	const operationsPerGoroutine = 2000
	const hotKeys = 10

	var wg sync.WaitGroup

	for g := 0; g < goroutines; g++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			for i := 0; i < operationsPerGoroutine; i++ {
				key := fmt.Sprintf("hot-key-%d", i%hotKeys)

				switch i % 4 {
				case 0:
					m.Put(key, fmt.Sprintf("value-%d", id))

				case 1:
					m.Get(key)

				case 2:
					m.Put(key, "updated")

				case 3:
					m.Get(key)
				}
			}
		}(g)
	}

	wg.Wait()

	t.Logf(
		"completed high-contention workload: %d goroutines × %d operations",
		goroutines,
		operationsPerGoroutine,
	)

	t.Logf("final map size: %d", m.Size())

	if m.Size() > hotKeys {
		t.Fatalf(
			"expected at most %d hot keys, got %d",
			hotKeys,
			m.Size(),
		)
	}
}
