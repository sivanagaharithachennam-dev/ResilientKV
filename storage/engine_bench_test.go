package storage

import (
	"fmt"
	"testing"
)

const benchmarkKeys = 1000

// createBenchmarkEngine creates an engine with a fixed dataset.
func createBenchmarkEngine(b *testing.B) *Engine {
	b.Helper()

	dataDir := b.TempDir()

	engine, err := New(dataDir)
	if err != nil {
		b.Fatalf("failed to create engine: %v", err)
	}

	// Keep all benchmark setup in memory so the benchmark
	// measures the requested operation rather than flushing.
	engine.memTableLimit = benchmarkKeys + 1

	for i := 0; i < benchmarkKeys; i++ {
		key := fmt.Sprintf("key-%04d", i)
		value := fmt.Sprintf("value-%04d", i)

		if err := engine.Put(key, value); err != nil {
			engine.Close()
			b.Fatalf("setup PUT failed: %v", err)
		}
	}

	return engine
}

// BenchmarkEnginePut measures PUT performance.
func BenchmarkEnginePut(b *testing.B) {
	dataDir := b.TempDir()

	engine, err := New(dataDir)
	if err != nil {
		b.Fatalf("failed to create engine: %v", err)
	}

	defer engine.Close()

	engine.memTableLimit = benchmarkKeys + 1

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf(
			"key-%04d",
			i%benchmarkKeys,
		)

		if err := engine.Put(key, "updated"); err != nil {
			b.Fatalf("PUT failed: %v", err)
		}
	}
}

// BenchmarkEngineGet measures GET performance.
func BenchmarkEngineGet(b *testing.B) {
	engine := createBenchmarkEngine(b)
	defer engine.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf(
			"key-%04d",
			i%benchmarkKeys,
		)

		_, exists := engine.Get(key)

		if !exists {
			b.Fatalf("key %s was not found", key)
		}
	}
}

// BenchmarkEngineDelete measures DELETE performance.
func BenchmarkEngineDelete(b *testing.B) {
	engine := createBenchmarkEngine(b)
	defer engine.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf(
			"key-%04d",
			i%benchmarkKeys,
		)

		if err := engine.Delete(key); err != nil {
			b.Fatalf("DELETE failed: %v", err)
		}
	}
}

// BenchmarkEngineMixed measures a mixed workload.
func BenchmarkEngineMixed(b *testing.B) {
	engine := createBenchmarkEngine(b)
	defer engine.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf(
			"key-%04d",
			i%benchmarkKeys,
		)

		switch i % 3 {
		case 0:
			if err := engine.Put(key, "updated"); err != nil {
				b.Fatalf("PUT failed: %v", err)
			}

		case 1:
			_, _ = engine.Get(key)

		case 2:
			if err := engine.Put(key, "updated-again"); err != nil {
				b.Fatalf("PUT failed: %v", err)
			}
		}
	}
}
