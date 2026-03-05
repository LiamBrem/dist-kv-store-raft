package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/liambrem/dist-kv-store-raft/kv"
	pb "github.com/liambrem/dist-kv-store-raft/proto"
	"github.com/liambrem/dist-kv-store-raft/raft"
	"google.golang.org/grpc"
)

func main() {
	// Parse command-line arguments
	nodeID := flag.Int("id", 0, "Node ID (e.g., 0, 1, 2)")
	port := flag.Int("port", 8000, "Port to listen on (e.g., 8000)")
	httpPort := flag.Int("http", 9000, "HTTP API port (e.g., 9000)")
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

	// Create KV store
	store := kv.NewKVStore(node)

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

	// Start HTTP API server
	go startHTTPServer(*httpPort, store)

	// Wait for interrupt signal to gracefully shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
	store.Stop()
	node.Stop()
	grpcServer.GracefulStop()
	log.Println("Shutdown complete")
}

func startHTTPServer(port int, store *kv.KVStore) {
	http.HandleFunc("/put", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := store.Put(req.Key, req.Value); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "key parameter required", http.StatusBadRequest)
			return
		}

		value, exists := store.Get(key)
		if !exists {
			http.Error(w, "key not found", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"key": key, "value": value})
	})

	http.HandleFunc("/keys", func(w http.ResponseWriter, r *http.Request) {
		keys := store.GetAllKeys()
		json.NewEncoder(w).Encode(map[string][]string{"keys": keys})
	})

	addr := fmt.Sprintf(":%d", port)
	log.Printf("HTTP API server listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
