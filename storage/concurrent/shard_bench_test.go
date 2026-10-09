package concurrent

import (
	"fmt"
	"testing"
)

func benchmarkConcurrentMapGetWithShards(b *testing.B, shardCount int) {
	m := NewConcurrentMap(shardCount)

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

func BenchmarkConcurrentMapGetShards8(b *testing.B) {
	benchmarkConcurrentMapGetWithShards(b, 8)
}

func BenchmarkConcurrentMapGetShards16(b *testing.B) {
	benchmarkConcurrentMapGetWithShards(b, 16)
}

func BenchmarkConcurrentMapGetShards32(b *testing.B) {
	benchmarkConcurrentMapGetWithShards(b, 32)
}

func BenchmarkConcurrentMapGetShards64(b *testing.B) {
	benchmarkConcurrentMapGetWithShards(b, 64)
}

func benchmarkConcurrentMapPutWithShards(b *testing.B, shardCount int) {
	m := NewConcurrentMap(shardCount)

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

func BenchmarkConcurrentMapPutShards8(b *testing.B) {
	benchmarkConcurrentMapPutWithShards(b, 8)
}

func BenchmarkConcurrentMapPutShards16(b *testing.B) {
	benchmarkConcurrentMapPutWithShards(b, 16)
}

func BenchmarkConcurrentMapPutShards32(b *testing.B) {
	benchmarkConcurrentMapPutWithShards(b, 32)
}

func BenchmarkConcurrentMapPutShards64(b *testing.B) {
	benchmarkConcurrentMapPutWithShards(b, 64)
}
