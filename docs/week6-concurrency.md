# Week 6 — Concurrency Testing, Stress Testing and Integration

## 1. Objective

Week 6 focused on extending the concurrent in-memory data structure
implemented during Week 5.

The main objectives were:

- High-concurrency stress testing
- High-contention workload testing
- Race detection
- Shard-count performance evaluation
- Storage-engine integration testing
- Full-project concurrency verification

## 2. High-Concurrency Stress Testing

The concurrent map was tested using:

- 100 concurrent goroutines
- 1000 operations per goroutine
- Concurrent Put, Get and Delete operations

The test completed successfully without data races.

## 3. High-Contention Testing

A high-contention workload was created using:

- 100 goroutines
- 2000 operations per goroutine
- 10 frequently accessed hot keys

The workload was designed to increase lock contention and test
concurrent access to the same keys.

The test completed successfully.

Observed final map size:

5 entries.

The final size is workload-dependent because multiple goroutines
continuously update the same small set of keys.

## 4. Race Detection

The concurrent package was tested using:

    go test -race ./storage/concurrent -v

The complete project was also tested using:

    go test -race ./...

All tests completed successfully without reported data races.

## 5. Sharded Concurrent Map

The implementation uses a sharded architecture.

Each shard contains:

- A map
- An RWMutex

Keys are distributed between shards using FNV-1a hashing.

Conceptually:

                ConcurrentMap
                      |
              Hash(key)
                      |
        +-------------+-------------+
        |             |             |
      Shard 1       Shard 2       Shard N
        |             |             |
      RWMutex       RWMutex       RWMutex
        |             |             |
       Map           Map           Map

The sharded design allows different keys to be protected by
different locks.

## 6. Shard-Count Benchmark

Shard-count benchmarking was prepared for:

- 8 shards
- 16 shards
- 32 shards
- 64 shards

The benchmark measures the effect of shard count on concurrent
Get and Put workloads.

Results should be interpreted based on the actual workload and
hardware environment rather than assuming that a larger number
of shards is always faster.

## 7. Storage Integration

A concurrent storage integration test was created to verify
concurrent access to the existing ResilientKV storage engine.

The test uses:

- 20 concurrent goroutines
- 100 operations per goroutine
- Concurrent Put and Get operations

The storage engine successfully handled the concurrent workload.

## 8. Architecture Decision

The existing MemTable already uses sync.RWMutex and has passed
race-enabled testing.

Therefore, the sharded ConcurrentMap was not directly substituted
for the MemTable.

This avoids unnecessary architectural complexity while preserving
the ConcurrentMap as a reusable concurrent data structure.

Potential future uses include:

- Concurrent indexes
- Metadata management
- Caching
- Client/session state
- In-memory distributed-system state

## 9. Week 6 Result

The concurrency layer was stress-tested and integrated with the
storage architecture.

The implementation passed:

- Normal unit tests
- Concurrent stress tests
- High-contention tests
- Race-enabled tests
- Storage integration tests

## 10. Week 5–6 Completion

The concurrent in-memory data-structure milestone is now complete.

Completed components:

- Sharded ConcurrentMap
- Concurrent Put/Get/Delete operations
- Unit tests
- Stress tests
- High-contention tests
- Race detection
- Performance benchmarks
- sync.Map comparison
- Storage integration testing

## 11. Next Milestone

The next major milestone is the distributed communication layer:

- gRPC service
- Client communication
- Request routing
- At-least-once semantics

This corresponds to the Week 7–8 distributed primitives milestone.
