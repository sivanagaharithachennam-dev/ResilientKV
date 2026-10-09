package concurrent

import (
	"fmt"
	"sync"
	"testing"
)

func BenchmarkSyncMapPut(b *testing.B) {
	var m sync.Map

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0

		for pb.Next() {
			key := fmt.Sprintf("key-%d", i%1000)
			m.Store(key, "value")
			i++
		}
	})
}

func BenchmarkSyncMapGet(b *testing.B) {
	var m sync.Map

	for i := 0; i < 1000; i++ {
		m.Store(fmt.Sprintf("key-%d", i), "value")
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0

		for pb.Next() {
			key := fmt.Sprintf("key-%d", i%1000)
			m.Load(key)
			i++
		}
	})
}

func BenchmarkSyncMapMixed(b *testing.B) {
	var m sync.Map

	for i := 0; i < 1000; i++ {
		m.Store(fmt.Sprintf("key-%d", i), "value")
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0

		for pb.Next() {
			key := fmt.Sprintf("key-%d", i%1000)

			switch i % 3 {
			case 0:
				m.Store(key, "value")
			case 1:
				m.Load(key)
			case 2:
				m.Delete(key)
			}

			i++
		}
	})
}
