package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/hashicorp/raft"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"

	raft_internal "kv-storage/internal/raft"
	"kv-storage/internal/server"
	"kv-storage/internal/store"
	pb "kv-storage/proto"
)

func main() {
	nodeID := flag.String("id", "node1", "Unique Node ID")
	grpcPort := flag.Int("grpc", 50051, "Port for gRPC (Client)")
	raftPort := flag.Int("raft", 60051, "Port for Raft (Cluster)")
	bootstrap := flag.Bool("bootstrap", false, "Bootstrap the cluster (use only for the first node)")

	flag.Parse()

	raftAddr := fmt.Sprintf("127.0.0.1:%d", *raftPort)
	httpAddr := fmt.Sprintf(":%d", *grpcPort)
	raftDir := fmt.Sprintf("./raft-data/%s", *nodeID)

	if err := os.MkdirAll(raftDir, 0755); err != nil {
		log.Fatalf("failed to create dir: %v", err)
	}
	metricsPort := *grpcPort + 1000
	metricsAddr := fmt.Sprintf(":%d", metricsPort)
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Printf("Metrics exposed on %s/metrics", metricsAddr)
		if err := http.ListenAndServe(metricsAddr, nil); err != nil {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	dbPath := filepath.Join(raftDir, "data.db")
	db, err := store.NewStore(dbPath)
	if err != nil {
		log.Fatalf("failed to create store: %v", err)
	}
	defer db.Close()

	fsm := &raft_internal.FSM{Store: db}
	r, err := raft_internal.SetupRaft(*nodeID, raftDir, raftAddr, fsm)
	if err != nil {
		log.Fatalf("failed to setup raft: %v", err)
	}

	if *bootstrap {
		log.Printf("Bootstrapping cluster as %s...", *nodeID)
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raft.ServerID(*nodeID),
					Address: raft.ServerAddress(raftAddr),
				},
			},
		}
		future := r.BootstrapCluster(configuration)
		if err := future.Error(); err != nil {
			log.Printf("Bootstrap error (maybe already done): %v", err)
		}
	}

	lis, err := net.Listen("tcp", httpAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	kvService := server.NewServer(db, r)
	pb.RegisterKVServiceServer(grpcServer, kvService)

	log.Printf("Node %s listening gRPC on %s, Raft on %s", *nodeID, httpAddr, raftAddr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
