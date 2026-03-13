# Distributed Key-Value Storage (RaftKV)

![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)
![gRPC](https://img.shields.io/badge/gRPC-API-244c5a?style=flat&logo=grpc)
![Raft](https://img.shields.io/badge/Consensus-Raft-blue)
![Kubernetes](https://img.shields.io/badge/Kubernetes-Ready-326ce5?style=flat&logo=kubernetes)
![License](https://img.shields.io/badge/License-MIT-green.svg)

A highly available, strongly consistent, distributed key-value storage system implemented in Go. 

This project leverages the **Raft consensus algorithm** to build a resilient cluster of nodes. It features a **custom-built, append-only storage engine** (inspired by Bitcask) for high-throughput writes, **gRPC** for efficient node-to-node and client-to-node communication, and built-in **Prometheus metrics** for observability.

## ✨ Key Features

*   **Strong Consistency & Fault Tolerance:** Uses HashiCorp's Raft implementation to replicate the state machine across multiple nodes. Survives network partitions and node failures.
*   **Custom Storage Engine:** 
    *   **Append-Only Log:** Ensures sequential disk writes for maximum performance.
    *   **In-Memory Index:** Maintains a hash map of offsets for **O(1)** read complexity.
    *   **Binary Encoding:** Custom binary format with CRC32 checksums for data integrity.
*   **gRPC API:** Strictly typed, high-performance communication interface for data operations (`Set`, `Get`) and cluster management (`Join`).
*   **Dynamic Clustering:** Supports adding new nodes on the fly without cluster downtime. Includes an auto-join mechanism.
*   **Kubernetes Ready:** Comes with a `StatefulSet` and Headless Service manifests for easy deployment in modern cloud environments.
*   **Observability:** Exposes Prometheus metrics (request counters, execution duration histograms) to monitor node health and latency.

## 🏗️ Architecture Design

1.  **Client Request:** A client sends a `Set(key, value)` gRPC request to the Leader node.
2.  **Raft Consensus:** The Leader appends the command to its Raft log and replicates it to Follower nodes.
3.  **Commit & Apply:** Once a quorum (majority) of nodes acknowledge the log, it is committed.
4.  **Storage Layer:** The Finite State Machine (FSM) applies the log to the custom `Store`, persisting the binary encoded entry to disk and updating the in-memory RAM index.

## 🚀 Getting Started

### Prerequisites
*   Go 1.25+
*   Docker & Kubernetes (optional, for cluster deployment)

### Local Setup

Clone the repository:
```bash
git clone https://github.com/morcux/kv-storage.git
cd kv-storage
go mod download
```

#### 1. Start the Leader (Bootstrap Node)
The first node must bootstrap the cluster.
```bash
go run cmd/server/main.go -id node1 -grpc 50051 -raft 60051
# Note: The server automatically bootstraps if it's the first node and no raft data exists.
```

#### 2. Start a Follower Node
Open a new terminal and start a second node on different ports:
```bash
go run cmd/server/main.go -id node2 -grpc 50052 -raft 60052
```

#### 3. Join the Cluster
Use the provided CLI tool to join `node2` to `node1`:
```bash
go run cmd/cli/main.go -cmd join -node-id node2 -raft-addr 127.0.0.1:60052 -server localhost:50051
```

### 🛠️ Interacting with the CLI

The project includes a CLI client (`cmd/cli`) to interact with the gRPC API.

**Set a key:**
```bash
go run cmd/cli/main.go -cmd set -key "username" -val "admin" -server localhost:50051
```

**Get a key:**
```bash
go run cmd/cli/main.go -cmd get -key "username" -server localhost:50051
```

## 🐳 Docker & Kubernetes

To deploy the highly available cluster in Kubernetes, apply the provided manifests. This will spin up a 3-node StatefulSet with persistent volumes.

```bash
# Build the image
docker build -t kv-storage:latest .

# Apply K8s manifests
kubectl apply -f deploy/manifests.yaml

# Check the pods
kubectl get pods -l app=kv-storage
```
*The pods will automatically discover and join the cluster based on the `POD_NAME` environment variables.*

## 📊 Monitoring

Metrics are exposed on the gRPC port + 1000 (e.g., `51051` for the default node).
You can scrape them using Prometheus:

```bash
curl http://localhost:51051/metrics
```
*Key metrics include: `kv_requests_total`, `kv_request_duration_seconds`.*

## 🧪 Testing

The project includes integration tests to ensure the storage engine and data encoding work correctly, including corruption recovery scenarios.

```bash
make test
# or
go test -v ./tests/...
```

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
