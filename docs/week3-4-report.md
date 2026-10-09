# ResilientKV Week 3–4 Report

## 1. Overview

During Week 3–4, the ResilientKV project focused on implementing
and testing the LSM-tree based storage engine.

The storage engine consists of:

- MemTable
- Write-Ahead Log (WAL)
- SSTable
- SSTable Index
- Compaction
- Recovery mechanism

The implementation was developed in Go.

---

## 2. Storage Architecture

The storage engine follows the following architecture:

Client
   |
   v
Storage Engine
   |
   +----------------+
   |                |
   v                v
  WAL            MemTable
                    |
                    v
                 SSTable
                    |
                    v
               Compaction

The WAL provides durability, the MemTable provides fast
in-memory access, SSTables provide persistent sorted storage,
and compaction merges older SSTables.

---

## 3. MemTable

The MemTable is an in-memory key-value structure implemented
using a Go map.

Operations supported:

- PUT
- GET
- DELETE
- Size
- Tombstone handling

A mutex-based synchronization mechanism is used to protect
concurrent access.

---

## 4. Write-Ahead Log

The WAL records operations before they are applied to the
MemTable.

Supported operations:

PUT key value
DELETE key

Each WAL append is followed by file synchronization to provide
durability.

The WAL is also used during startup recovery to reconstruct the
active MemTable.

---

## 5. SSTable

When the MemTable is flushed, its contents are written to an
SSTable.

SSTable entries are stored in sorted key order.

The current implementation supports:

- Key-value entries
- Tombstones
- Sequential reads
- Key lookup
- SSTable indexing

---

## 6. Compaction

Compaction merges multiple SSTables into a single SSTable.

The compaction process:

1. Finds existing SSTables.
2. Reads their entries.
3. Resolves newer values over older values.
4. Preserves deletion tombstones.
5. Writes a new SSTable.
6. Removes obsolete SSTables.

Compaction was tested for both updated values and deleted keys.

---

## 7. Recovery

The storage engine reads the WAL during startup.

The recovery process replays:

PUT operations

and

DELETE operations

to reconstruct the active MemTable.

Persistent SSTables are also retained for historical storage.

---

## 8. Stress Testing

An LSM stress test was implemented using:

- 1,000 keys
- Multiple MemTable flushes
- Key updates
- Key deletions
- SSTable compaction
- Post-compaction verification

The stress test successfully verified:

- Inserted keys
- Updated values
- Deleted keys
- Tombstone behavior
- Compaction correctness
- Data retrieval after compaction

The stress test also passed with the Go race detector.

---

## 9. Performance Benchmark

Command:

go test ./storage -bench=. -benchmem -benchtime=1s

Environment:

CPU:
13th Gen Intel(R) Core(TM) i5-1340P

Architecture:
amd64

### Results

| Operation | ns/op | B/op | allocs/op |
|-----------|------:|-----:|----------:|
| PUT | 2,406,378 | 283 | 4 |
| GET | 131.7 | 13 | 1 |
| DELETE | 2,171,113 | 44 | 3 |
| Mixed | 1,433,573 | 53 | 3 |

### Interpretation

GET operations are significantly faster because the benchmark
retrieves data directly from the in-memory MemTable.

PUT and DELETE operations have higher latency because the current
WAL implementation performs synchronous file synchronization
for every operation.

Therefore, the write latency includes the durability cost of
WAL persistence.

---

## 10. Testing Status

The following tests have been completed:

- MemTable unit tests
- WAL unit tests
- SSTable unit tests
- SSTable index tests
- Storage engine tests
- Persistent SSTable ID test
- Compaction tests
- Tombstone recovery tests
- LSM stress test
- Race detector testing
- Storage performance benchmark

All current storage tests and race-detector tests pass.

---

## 11. Week 3–4 Status

The following Week 3–4 components have been implemented:

[✓] MemTable
[✓] WAL
[✓] SSTable
[✓] SSTable index
[✓] Compaction
[✓] Recovery
[✓] Stress testing
[✓] Race testing
[✓] ResilientKV benchmark

Remaining Week 4 task:

[ ] Benchmark ResilientKV against RocksDB

---

## 12. Future Improvements

Potential improvements include:

- WAL rotation
- WAL truncation after safe checkpointing
- Background MemTable flushing
- Background compaction
- Bloom filters
- Block-based SSTable format
- More efficient SSTable indexing
- Batch writes
- Configurable WAL durability
- Improved read path
- RocksDB performance comparison

---

## 13. Conclusion

The Week 3–4 implementation established the core persistent
storage layer of ResilientKV.

The system now supports in-memory writes, durable WAL logging,
SSTable persistence, recovery, tombstone handling, compaction,
stress testing, and performance benchmarking.

The next major task is to perform a controlled performance
comparison against RocksDB before proceeding to the next
project milestone.

---

## 14. Initial RocksDB Benchmark

RocksDB version 8.9.1 was installed and benchmarked using its
native C++ API.

Environment:

- CPU: 13th Gen Intel(R) Core(TM) i5-1340P
- Architecture: amd64
- Compiler: g++ 13.3.0
- RocksDB: 8.9.1
- Dataset: 1,000 keys
- Operations: 10,000 per benchmark

Initial results:

| Operation | RocksDB |
|-----------|--------:|
| PUT | 5,511.88 ns/op |
| GET | 1,618.05 ns/op |
| DELETE | 7,790.22 ns/op |
| Mixed | 5,449.83 ns/op |

### Methodology Limitation

The initial RocksDB benchmark uses RocksDB's default
WriteOptions, whereas ResilientKV currently performs
synchronous WAL file synchronization for every write.

Consequently, these measurements should be treated as
baseline observations rather than a direct apples-to-apples
comparison.

A controlled comparison with explicitly matched durability
settings is required before making performance conclusions.

---

## 15. Benchmark Comparison Plan

The final performance evaluation will control the following
variables:

- Same CPU and operating environment
- Same number of keys
- Same operation count
- Same key/value sizes
- Same workload distribution
- Comparable durability configuration
- Same benchmark warm-up methodology
- Multiple benchmark repetitions

The final report will present the raw measurements and explain
the configuration used for each system.

---

## 14. Initial RocksDB Benchmark

RocksDB version 8.9.1 was installed and benchmarked using its
native C++ API.

Environment:

- CPU: 13th Gen Intel(R) Core(TM) i5-1340P
- Architecture: amd64
- Compiler: g++ 13.3.0
- RocksDB: 8.9.1
- Dataset: 1,000 keys
- Operations: 10,000 per benchmark

Initial results:

| Operation | RocksDB |
|-----------|--------:|
| PUT | 5,511.88 ns/op |
| GET | 1,618.05 ns/op |
| DELETE | 7,790.22 ns/op |
| Mixed | 5,449.83 ns/op |

### Methodology Limitation

The initial RocksDB benchmark uses RocksDB's default
WriteOptions, whereas ResilientKV currently performs
synchronous WAL file synchronization for every write.

Consequently, these measurements should be treated as
baseline observations rather than a direct apples-to-apples
comparison.

A controlled comparison with explicitly matched durability
settings is required before making performance conclusions.

---

## 15. Benchmark Comparison Plan

The final performance evaluation will control the following
variables:

- Same CPU and operating environment
- Same number of keys
- Same operation count
- Same key/value sizes
- Same workload distribution
- Comparable durability configuration
- Same benchmark warm-up methodology
- Multiple benchmark repetitions

The final report will present the raw measurements and explain
the configuration used for each system.
