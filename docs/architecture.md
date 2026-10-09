# ResilientKV Architecture

## System Objective

ResilientKV is a distributed key-value store designed to provide
reliable storage and fault tolerance using Raft consensus.

## Major Components

1. Client
2. gRPC Service
3. Raft Consensus Layer
4. LSM-Tree Storage Engine
5. Kubernetes Deployment
6. Monitoring
7. Fault Injection
8. Benchmarking

## Initial Architecture

Client
   |
   v
gRPC Service
   |
   v
Raft Consensus
   |
   v
LSM Storage Engine
   |
   +-- MemTable
   +-- WAL
   +-- SSTable
