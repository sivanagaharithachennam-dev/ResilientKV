package concurrent

import (
	"fmt"
	"sync"
	"testing"
)

func BenchmarkConcurrentMapPut(b *testing.B) {
	m := NewConcurrentMap(32)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0

		for pb.Next() {
			key := fmt.Sprintf("key-%d", i%1000)
			m.Put(key, "value")
			i++
		}
	})
}

func BenchmarkConcurrentMapGet(b *testing.B) {
	m := NewConcurrentMap(32)

	for i := 0; i < 1000; i++ {
		m.Put(fmt.Sprintf("key-%d", i), "value")
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0

		for pb.Next() {
			key := fmt.Sprintf("key-%d", i%1000)
			m.Get(key)
			i++
		}
	})
}

func BenchmarkConcurrentMapMixed(b *testing.B) {
	m := NewConcurrentMap(32)

	for i := 0; i < 1000; i++ {
		m.Put(fmt.Sprintf("key-%d", i), "value")
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0

		for pb.Next() {
			key := fmt.Sprintf("key-%d", i%1000)

			switch i % 3 {
			case 0:
				m.Put(key, "value")
			case 1:
				m.Get(key)
			case 2:
				m.Delete(key)
			}

			i++
		}
	})
}

func BenchmarkConcurrentMapExplicitGoroutines(b *testing.B) {
	m := NewConcurrentMap(32)

	b.ResetTimer()

	var wg sync.WaitGroup

	for i := 0; i < b.N; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			key := fmt.Sprintf("key-%d", id%1000)
			m.Put(key, "value")
			m.Get(key)
		}(i)
	}

	wg.Wait()
}
