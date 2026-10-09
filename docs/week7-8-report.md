# ResilientKV — Week 7–8 Report
## Distributed Communication and Testing

### 1. Objective
Implement a gRPC communication layer connecting clients to the
ResilientKV storage engine, and test basic distributed-service behavior.

### 2. Technology
- Go
- gRPC
- Protocol Buffers
- Existing ResilientKV storage engine

### 3. Implemented Components
- Protocol Buffers API for PUT, GET, and DELETE
- gRPC server
- gRPC client
- Storage-engine integration
- Request IDs for duplicate-request detection
- Basic deterministic key-to-node routing component

### 4. Functional Testing
The gRPC client successfully performed PUT, GET, and DELETE operations.
A GET after DELETE confirmed that the deleted key was not found.

### 5. Duplicate-Request Test
The client sent the same PUT request ID twice.
Both calls returned success, and a subsequent GET found the expected value.

Limitation: this test demonstrates duplicate-request handling at the
service interface. It does not yet prove durable deduplication across
server restarts or replicated exactly-once execution.

### 6. Routing Test
The routing unit test passed and mapped the key "product" to a node
address. This validates the basic routing function only. Routing
requests across live nodes and selecting a Raft leader remain future work.

### 7. Server-Unavailability Test
When the gRPC server was stopped, the client received an RPC error
with status Unavailable and connection refused. This confirms that
the client reports this connection failure. Automatic failover is not
implemented by this test.

### 8. Automated Test Results
Commands executed:

    go test ./...
    go test -race ./...

All listed Go packages passed. No race-detector warnings were reported.

### 9. Current Limitations
- Raft consensus and leader election are not implemented yet.
- The routing component is not yet integrated with live multi-node RPC.
- Request deduplication is in memory and is not durable across restarts.
- Kubernetes deployment and production observability are future milestones.

### 10. Conclusion
Weeks 7–8 established and tested the basic gRPC communication layer.
The next milestone is implementing Raft consensus, including elections,
log replication, commit handling, and safety tests.
