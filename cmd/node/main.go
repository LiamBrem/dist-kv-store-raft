package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	pb "github.com/liambrem/dist-kv-store-raft/proto"
	"github.com/liambrem/dist-kv-store-raft/raft"
	"google.golang.org/grpc"
)

func main() {
	// Parse command-line arguments
	nodeID := flag.Int("id", 0, "Node ID (e.g., 0, 1, 2)")
	port := flag.Int("port", 8000, "Port to listen on (e.g., 8000)")
	peersStr := flag.String("peers", "", "Comma-separated peer addresses (e.g., localhost:8001,localhost:8002)")
	flag.Parse()

	// Validate arguments
	if *nodeID < 0 {
		log.Fatal("Node ID must be >= 0")
	}

	// Parse peer addresses
	var peers []string
	if *peersStr != "" {
		peers = strings.Split(*peersStr, ",")
		// Trim whitespace
		for i := range peers {
			peers[i] = strings.TrimSpace(peers[i])
		}
	}

	log.Printf("Starting Raft node %d on port %d", *nodeID, *port)
	log.Printf("Peers: %v", peers)

	// Create Raft node
	node := raft.NewRaftNode(*nodeID, peers)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterRaftServiceServer(grpcServer, node)

	// Start listening
	addr := fmt.Sprintf(":%d", *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}

	// Start the Raft node
	node.Start()
	log.Printf("Raft node %d started", *nodeID)

	// Start gRPC server in a goroutine
	go func() {
		log.Printf("gRPC server listening on %s", addr)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
	node.Stop()
	grpcServer.GracefulStop()
	log.Println("Shutdown complete")
}
