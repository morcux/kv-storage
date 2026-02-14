package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "kv-storage/proto"
)

func main() {
	serverAddr := flag.String("server", "localhost:50051", "Address of the gRPC server (Leader)")
	cmd := flag.String("cmd", "get", "Command to execute: get, set, join")

	key := flag.String("key", "", "Key for set/get")
	val := flag.String("val", "", "Value for set")

	nodeID := flag.String("node-id", "", "Node ID to join (e.g., node2)")
	raftAddr := flag.String("raft-addr", "", "Raft address of the new node (e.g., 127.0.0.1:60052)")

	flag.Parse()

	conn, err := grpc.NewClient(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewKVServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	switch *cmd {
	case "set":
		if *key == "" || *val == "" {
			log.Fatal("missing -key or -val for set command")
		}
		_, err := client.Set(ctx, &pb.SetRequest{Key: *key, Value: []byte(*val)})
		if err != nil {
			log.Fatalf("Set failed: %v", err)
		}
		fmt.Println("Set successful")

	case "get":
		if *key == "" {
			log.Fatal("missing -key for get command")
		}
		resp, err := client.Get(ctx, &pb.GetRequest{Key: *key})
		if err != nil {
			log.Fatalf("Get failed: %v", err)
		}
		fmt.Printf("Value: %s\n", resp.Value)

	case "join":
		if *nodeID == "" || *raftAddr == "" {
			log.Fatal("missing -node-id or -raft-addr for join command")
		}
		_, err := client.Join(ctx, &pb.JoinRequest{
			NodeID:      *nodeID,
			RaftAddress: *raftAddr,
		})
		if err != nil {
			log.Fatalf("Join failed: %v", err)
		}
		fmt.Printf("Node %s joined successfully via %s\n", *nodeID, *serverAddr)

	default:
		log.Fatalf("unknown command: %s", *cmd)
	}
}
