# Week 2 - LSM-Tree Storage Engine

## Objective

Implement the storage layer of ResilientKV using an LSM-tree architecture.

## LSM Components

1. MemTable
2. Write-Ahead Log (WAL)
3. SSTable
4. Compaction

## Week 2 Progress

### MemTable

Implemented an in-memory key-value structure using a Go map.

Operations:

- Put
- Get
- Delete
- Size

### Concurrency

The MemTable uses `sync.RWMutex` to support concurrent access safely.

### Testing

Unit tests verify:

- Insert
- Read
- Update
- Delete
- Size

Race detection is performed using:

`go test -race`

### Benchmarking

MemTable Put and Get operations are benchmarked using Go's benchmark framework.

## Architecture

Client
    |
    v
MemTable
    |
    v
WAL
    |
    v
SSTable
    |
    v
Compaction
