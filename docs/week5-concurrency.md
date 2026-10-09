# Week 5 — Concurrent In-Memory Data Structures

## 1. Objective

The objective of Week 5 was to implement and evaluate concurrent
in-memory data structures for the ResilientKV distributed key-value store.

The implementation focuses on safe concurrent access using Go
goroutines and synchronization primitives.

## 2. Concurrent Map

A sharded concurrent map was implemented.

Each shard contains:

- A Go map
- An RWMutex
- Independent locking

The key is hashed using FNV-1a to select a shard.

Architecture:

Client Goroutines
       |
       v
ConcurrentMap
       |
   Hash Key
       |
       v
+------+------+------+
| Shard | Shard | ...|
|  RW   |  RW   |    |
|  Map  |  Map  |    |
+------+------+------+

This reduces contention compared with protecting the entire map
using a single global lock.

## 3. Operations Implemented

The concurrent map supports:

- Put
- Get
- Delete
- Size

## 4. Testing

Unit tests were created for:

- Insert and retrieve
- Update
- Delete
- Size
- Concurrent access

All tests passed.

## 5. Race Detection

The following command was executed:

    go test -race ./storage/concurrent

Result:

    PASS

The complete project was also tested using:

    go test -race ./...

No data-race warnings were reported.

## 6. Benchmark Environment

Operating system:

Linux / WSL2

Architecture:

amd64

CPU:

13th Gen Intel(R) Core(TM) i5-1340P

Benchmark duration:

3 seconds per benchmark

## 7. ConcurrentMap Benchmark

| Operation | ns/op | B/op | allocs/op |
|---|---:|---:|---:|
| Put | 163.4 | 13 | 1 |
| Get | 35.16 | 13 | 1 |
| Mixed | 157.5 | 13 | 1 |
| Explicit Goroutines | 395.9 | 62 | 3 |

## 8. Comparison with Go sync.Map

| Operation | ResilientKV ConcurrentMap | Go sync.Map |
|---|---:|---:|
| Put | 163.4 ns/op | 129.5 ns/op |
| Get | 35.16 ns/op | 32.85 ns/op |
| Mixed | 157.5 ns/op | 106.9 ns/op |

Allocation results:

| Operation | ConcurrentMap | sync.Map |
|---|---:|---:|
| Put | 13 B/op, 1 alloc | 78 B/op, 3 allocs |
| Get | 13 B/op, 1 alloc | 13 B/op, 1 alloc |
| Mixed | 13 B/op, 1 alloc | 38 B/op, 2 allocs |

These measurements are workload- and environment-dependent and
should not be interpreted as a universal performance ranking.

## 9. Race Benchmark

Race-enabled benchmarks were also executed.

The race detector increased execution time, as expected because
additional instrumentation is performed.

The benchmark completed successfully without data-race warnings.

## 10. Result

A sharded concurrent map was successfully implemented for ResilientKV.

The implementation supports concurrent Put, Get, Delete, and Size
operations and passed both normal and race-enabled testing.

## 11. Week 5 Status

Week 5 objectives completed:

- Concurrent data structure implemented
- Concurrent tests implemented
- Race detection completed
- Performance benchmarks completed
- Comparison with Go sync.Map completed

## 12. Future Work

Week 6 will extend the concurrency work with additional benchmarking,
stress testing, and integration considerations for the storage engine
and distributed components.
