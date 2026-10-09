# ResilientKV
## Distributed Key-Value Store with Raft Consensus

**Project:** M.Tech Trimester Project  
**Language:** Go  
**Communication:** gRPC and Protocol Buffers  
**Storage:** LSM-tree-style engine  
**Consensus:** Raft  

## 1. Project Overview

ResilientKV is an educational distributed key-value storage system developed using Go. It combines persistent storage, concurrent data structures, gRPC communication, and Raft-based leader election and replicated-write handling.

The project demonstrates storage-engine design, recovery-related mechanisms, distributed coordination, and automated testing.

## 2. Objectives

- Implement PUT, GET, and DELETE operations.
- Develop a Write-Ahead Log (WAL).
- Implement memtable and SSTable storage.
- Integrate storage recovery and compaction.
- Develop and test concurrent data structures.
- Provide client-server communication using gRPC.
- Implement Raft election, voting, log replication, and commit handling.
- Test storage and distributed-system components.
- Prepare for performance benchmarking, fault injection, monitoring, and Kubernetes deployment.

## 3. Technology Stack

- Go
- gRPC
- Protocol Buffers
- Raft consensus
- WAL, memtable, SSTables, and compaction
- Go unit and integration tests
- Go race detector
- Ubuntu on WSL
- Git and GitHub

## 4. System Architecture

    KV Client
        |
        v
    gRPC Service: PUT / GET / DELETE
        |
        v
    Raft Coordination
    Election / Replication / Commit
        |
        v
    Storage Engine
        |
        +-- WAL
        +-- Memtable
        +-- SSTables
        +-- Compaction

## 5. Twelve-Week Development Plan

### Weeks 1-2: Requirements and Project Setup
- Define project objectives and architecture.
- Set up the Go environment and repository.
- Establish the module structure and build workflow.
- Plan testing and development milestones.

**Status:** Core project structure and Go implementation established.

### Weeks 3-4: LSM-Tree Storage Engine
- Implement the memtable.
- Develop WAL writing and recovery-related behavior.
- Implement SSTable storage and lookup.
- Integrate storage operations and tests.

**Status:** WAL, memtable, SSTable, and storage-engine components implemented and tested. A formal RocksDB comparison requires separate benchmark evidence.

### Weeks 5-6: Concurrent Data Structures
- Implement concurrent storage utilities.
- Test concurrent access.
- Run race-related checks.
- Prepare benchmarks against standard-library structures.

**Status:** Concurrent storage utilities and tests implemented. Comparative benchmark results should be documented separately.

### Weeks 7-8: Distributed Primitives and gRPC
- Define Protocol Buffer messages and gRPC services.
- Implement client-server communication.
- Support key-value RPC operations.
- Test repeated requests and request identifiers.
- Develop routing components.

**Status:** gRPC service, client, key-value operations, request-ID handling, and routing components implemented and tested.

### Weeks 9-10: Raft Consensus
- Implement Raft node state and voting.
- Implement leader election.
- Handle AppendEntries requests.
- Replicate log entries to followers.
- Track commit indexes and apply committed commands.
- Test majority-related behavior.

**Status:** Raft election, voting, replication, and commit-handling components implemented. Automated tests cover single-node writes, follower write rejection, insufficient-majority behavior, deletion, and gRPC follower replication.

### Weeks 11-12: Deployment, Monitoring, and Fault Testing
- Prepare Kubernetes manifests and a Helm chart.
- Add Prometheus metrics and Grafana dashboards.
- Develop a Jepsen-style fault-injection harness.
- Run multi-node and failure-recovery experiments.
- Compare performance with established systems.
- Prepare architecture documentation and the final demonstration.

**Status:** Deployment and evaluation tasks must be marked complete only after their artifacts and results are verified.

## 6. Implemented Components

### Storage Engine
Provides key-value operations and integrates persistent storage components.

### Write-Ahead Log
Records updates and supports recovery-related behavior.

### Memtable
Maintains recent key-value updates in memory.

### SSTable
Provides disk-based table storage and lookup.

### Compaction
Reorganizes stored data and handles obsolete records.

### Concurrent Data Structures
Provides concurrent storage utilities and associated tests.

### gRPC
Provides client-server communication and Raft-related RPC methods using Protocol Buffers.

### Raft
Implements election, voting, log replication, and commit-handling components.

### Replicated Writes
The write path coordinates replication, checks majority acknowledgements, advances commit indexes, and applies committed commands to storage.

## 7. Project Structure

    ResilientKV/
    ├── cmd/
    │   ├── client/
    │   ├── raft-election-client/
    │   ├── raft-vote-client/
    │   └── server/
    ├── grpc/
    │   ├── proto/
    │   └── routing/
    ├── raft/
    ├── storage/
    │   ├── memtable/
    │   ├── sstable/
    │   ├── wal/
    │   ├── compaction/
    │   └── concurrent/
    ├── README.md
    └── go.mod

## 8. Build and Test

Build all packages:

    go build ./...

Run the complete test suite:

    go test -count=1 ./...

Run the gRPC follower-replication integration test:

    go test -count=1 -v ./grpc -run TestRaftPutReplicatesToFollowerOverGRPC

Run the race detector:

    go test -race ./...

These commands were used during project verification. Record fresh results when preparing the final report.

## 9. Verified Demonstration Results

The client successfully:
1. Sent a PUT request.
2. Retried the same request using the same request ID.
3. Retrieved the stored value.

Observed result:

    First PUT: success=true
    Retry PUT: success=true
    GET: key=product value=laptop found=true

The complete Go test suite passed during verification. The dedicated gRPC follower-replication integration test also passed.

The election client reported a successful election in a demonstration run. A successful election response alone does not prove that all three live nodes participated; verify node configuration and voting evidence separately.

## 10. Running the Demo

Start a server:

    RAFT_NODE_ID=node1 RESILIENTKV_PORT=50051 RESILIENTKV_DATA_DIR=/tmp/resilientkv-demo-data go run ./cmd/server

In a second terminal, run the client:

    cd /mnt/c/Users/DELL/Documents/OS_Project/ResilientKV
    go run ./cmd/client

Run the replication integration test:

    go test -count=1 -v ./grpc -run TestRaftPutReplicatesToFollowerOverGRPC

Run the complete test suite:

    go test -count=1 ./...

## 11. Original Trimester Project Deliverables

The original specification calls for the following deliverables.

1. Distributed key-value store source code.
2. LSM-tree storage engine as a separate reusable library.
3. Raft implementation with safety tests.
4. Helm chart for Kubernetes deployment.
5. Prometheus and Grafana dashboards.
6. Jepsen-style test harness with documented results.
7. Architecture document of approximately 10-15 pages.
8. Live demonstration on a multi-node Kubernetes cluster.
9. Performance benchmark report against systems such as RocksDB, etcd, or TiKV.

The core Go storage, gRPC, and Raft components have been implemented and tested. The remaining deliverables need separate verification and evidence before being marked complete.

## 12. Remaining Evaluation Checklist

- [ ] Linux profiling baseline using perf and/or bpftrace.
- [ ] Reproducible RocksDB benchmark comparison.
- [ ] Benchmark comparison against standard-library concurrent structures.
- [ ] Documented client-routing and at-least-once delivery behavior.
- [ ] Broader Raft safety and multi-node failure testing.
- [ ] Kubernetes deployment manifests and multi-node deployment.
- [ ] Helm chart.
- [ ] Prometheus instrumentation.
- [ ] Importable Grafana dashboards.
- [ ] Jepsen-style fault-injection harness.
- [ ] Documented fault-testing results.
- [ ] Separate 10-15-page architecture document.
- [ ] Performance report against appropriate reference systems.
- [ ] Live demonstration on a multi-node Kubernetes cluster.

## 13. Limitations

ResilientKV is an educational implementation and is not yet a production-ready distributed database.

Further engineering includes:
- Durable Raft term, voting, and log metadata.
- Robust election timeout handling and automatic leader re-election.
- Reliable follower catch-up after disconnection.
- Broader crash-recovery and network-partition testing.
- Reproducible throughput and latency benchmarks.
- Kubernetes deployment and operational tooling.
- Monitoring and fault-injection testing.
- Production-grade security and hardening.

## 14. Future Enhancements

- Persist Raft metadata and improve restart recovery.
- Improve election timeout handling and leader re-election.
- Implement reliable follower catch-up.
- Expand safety and fault-injection tests.
- Benchmark throughput, latency, and resource utilization.
- Compare against established storage systems.
- Add Prometheus metrics and Grafana dashboards.
- Deploy using Kubernetes and Helm.
- Extract the storage engine into a reusable library.

## 15. Conclusion

ResilientKV demonstrates the integration of persistent key-value storage, concurrent data structures, gRPC communication, and Raft-based distributed coordination.

The implementation includes storage-engine components, RPC services, election handling, log replication, and commit-related behavior. Automated tests validate the implemented storage and distributed components.

The next stage is to complete and document the remaining deployment, observability, benchmarking, and fault-injection deliverables to evaluate behavior under realistic distributed-system conditions.
