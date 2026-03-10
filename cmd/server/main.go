package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"kv-storage/internal/config"
	raft_internal "kv-storage/internal/raft"
	"kv-storage/internal/server"
	"kv-storage/internal/store"
	pb "kv-storage/proto"
)

func checkEmptyDisk(raftDir string) bool {
	_, err := os.Stat(raftDir)
	return os.IsNotExist(err)
}

func joinCluster(seedAddr, nodeID, raftAdvertiseAddr string) {
	log.Printf("Starting auto-join process to Seed Node: %s", seedAddr)

	for i := 0; i < 10; i++ {
		time.Sleep(3 * time.Second)

		conn, err := grpc.NewClient(seedAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("Auto-join dial error: %v", err)
			continue
		}

		client := pb.NewKVServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

		_, err = client.Join(ctx, &pb.JoinRequest{
			NodeID:      nodeID,
			RaftAddress: raftAdvertiseAddr,
		})
		cancel()
		conn.Close()

		if err == nil {
			log.Printf("Successfully joined the cluster via %s", seedAddr)
			return
		}

		log.Printf("Failed to join cluster (attempt %d/10): %v", i+1, err)
	}
	log.Fatalf("Could not join the cluster after multiple attempts")
}

func main() {
	cfg := config.LoadServerConfig()

	nodeID := flag.String("id", cfg.NodeID, "Unique Node ID")
	grpcPort := flag.Int("grpc", cfg.GRPCPort, "Port for gRPC (Client)")
	raftPort := flag.Int("raft", cfg.RaftPort, "Port for Raft (Cluster)")

	flag.Parse()

	httpBindAddr := fmt.Sprintf("0.0.0.0:%d", *grpcPort)
	raftBindAddr := fmt.Sprintf("0.0.0.0:%d", *raftPort)

	raftAdvertiseAddr := fmt.Sprintf("%s.my-raft-db-service.default.svc.cluster.local:%d", *nodeID, *raftPort)

	raftDir := fmt.Sprintf("./raft-data/%s", *nodeID)
	isNewCluster := checkEmptyDisk(raftDir)

	if err := os.MkdirAll(raftDir, 0755); err != nil {
		log.Fatalf("failed to create dir: %v", err)
	}
	metricsPort := *grpcPort + 1000
	metricsAddr := fmt.Sprintf("0.0.0.0:%d", metricsPort)
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
	r, err := raft_internal.SetupRaft(*nodeID, raftDir, raftBindAddr, raftAdvertiseAddr, fsm)
	if err != nil {
		log.Fatalf("failed to setup raft: %v", err)
	}

	if isNewCluster {
		if cfg.NodeID == cfg.SeedNodeID {
			log.Printf("Bootstrapping cluster as %s...", *nodeID)
			configuration := raft.Configuration{
				Servers: []raft.Server{
					{
						ID:      raft.ServerID(*nodeID),
						Address: raft.ServerAddress(raftAdvertiseAddr),
					},
				},
			}
			future := r.BootstrapCluster(configuration)
			if err := future.Error(); err != nil {
				log.Printf("Bootstrap error (maybe already done): %v", err)
			}
		} else {
			go joinCluster(cfg.SeedNodeAddr, *nodeID, raftAdvertiseAddr)
		}
	}

	lis, err := net.Listen("tcp", httpBindAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	kvService := server.NewServer(db, r)
	pb.RegisterKVServiceServer(grpcServer, kvService)

	log.Printf("Node %s listening gRPC on %s, Raft on %s", *nodeID, httpBindAddr, raftBindAddr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
