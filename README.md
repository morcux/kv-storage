# Distributed Key-Value Storage

A distributed, consistent key-value storage system implemented in Go, leveraging the [Raft consensus algorithm](https://raft.github.io/) for fault tolerance and data replication. This project creates a resilient cluster where data is replicated across multiple nodes, ensuring availability even in the event of node failures.

## Features

*   **Raft Consensus**: Uses the HashiCorp Raft implementation to ensure strong consistency across the cluster.
*   **gRPC API**: generic interface for client interactions (`Set`, `Get`) and cluster management (`Join`).
*   **Custom Storage Engine**: Implements a persistent append-only log storage with an in-memory index for fast lookups.
*   **Cluster Membership**: Dynamic node addition allows scaling the cluster without downtime.
*   **Metrics**: Prometheus metrics exposed for monitoring node health and performance.
*   **CLI Tool**: Included command-line interface for easy interaction and testing.

## Getting Started

### Prerequisites

*   Go 1.21 or higher
*   Make (optional, for build commands)

### Installation

Clone the repository and download dependencies:

```bash
git clone https://github.com/morcux/kv-storage.git
cd kv-storage
go mod download
```

## Usage

### 1. Start the Leader Node

The first node must be started in bootstrap mode to initialize the Raft cluster.

```bash
# Start node1 on gRPC port 50051 and Raft port 60051
go run cmd/server/main.go -id node1 -grpc 50051 -raft 60051 -bootstrap
```

*   `-id`: Unique identifier for the node.
*   `-bootstrap`: Initializes the cluster (only for the first node).
*   Metrics will be available at `http://localhost:60051/metrics`.

### 2. Start Additional Nodes

Start other nodes without the bootstrap flag. To make them part of the cluster, they must be "joined" via the leader.

```bash
# Start node2 on gRPC port 50052 and Raft port 60052
go run cmd/server/main.go -id node2 -grpc 50052 -raft 60052
```

### 3. Join Nodes to the Cluster

Use the CLI to tell the leader (running on port 50051) to add the new node (node2) to the cluster.

```bash
go run cmd/cli/main.go -cmd join -node-id node2 -raft-addr 127.0.0.1:60052 -server localhost:50051
```

*   `-raft-addr`: The Raft address of the *new* node you are adding.
*   `-server`: The gRPC address of the *leader* node.

### 4. Key-Value Operations

You can now read and write data to the cluster.

**Set a Value:**
```bash
go run cmd/cli/main.go -cmd set -key my-key -val "Hello Raft" -server localhost:50051
```

**Get a Value:**
```bash
go run cmd/cli/main.go -cmd get -key my-key -server localhost:50051
```

*   Note: While writes (Set) must be handled by the leader, this implementation forwards requests or requires clients to connect to the leader.

## Architecture

*   **`cmd/server`**: The main entry point for the storage node. Initializes Raft, the gRPC server, and the storage engine.
*   **`cmd/cli`**: A utility client for interacting with the cluster.
*   **`internal/raft`**: Handles Raft node setup, FSM (Finite State Machine) application, and snapshotting.
*   **`internal/store`**: A custom thread-safe storage engine. It uses an append-only file for persistence and maintains an in-memory hash map index for O(1) reads.
*   **`proto`**: Protocol Buffer definitions for the gRPC service.

## API Reference

The service exposes the following gRPC methods:

*   `Set(key, value)`: Replicates a command to store a value.
*   `Get(key)`: Retrieves a value from the local state machine.
*   `Join(nodeID, raftAddress)`: Adds a new peer to the Raft configuration.

## License

MIT
