package concurrent

import (
	"fmt"
	"sync"
	"testing"
)

func TestConcurrentMapHighConcurrency(t *testing.T) {
	m := NewConcurrentMap(32)

	const goroutines = 100
	const operationsPerGoroutine = 1000

	var wg sync.WaitGroup

	for g := 0; g < goroutines; g++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			for i := 0; i < operationsPerGoroutine; i++ {
				key := fmt.Sprintf("goroutine-%d-key-%d", id, i%100)

				switch i % 3 {
				case 0:
					m.Put(key, "value")

				case 1:
					m.Get(key)

				case 2:
					m.Delete(key)
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

	t.Logf("final map size: %d", m.Size())
}
