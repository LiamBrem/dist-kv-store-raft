package tests

import (
	"fmt"
	"net"
	"testing"
	"time"

	pb "github.com/liambrem/dist-kv-store-raft/proto"
	"github.com/liambrem/dist-kv-store-raft/raft"
	"google.golang.org/grpc"
)

// TestCluster manages a cluster of Raft nodes for testing
type TestCluster struct {
	nodes     []*raft.RaftNode
	servers   []*grpc.Server
	listeners []net.Listener
	ports     []int
	t         *testing.T
}

// NewTestCluster creates a cluster of n Raft nodes
func NewTestCluster(t *testing.T, n int) *TestCluster {
	cluster := &TestCluster{
		nodes:     make([]*raft.RaftNode, n),
		servers:   make([]*grpc.Server, n),
		listeners: make([]net.Listener, n),
		ports:     make([]int, n),
		t:         t,
	}

	// Allocate ports
	basePort := 10000
	for i := 0; i < n; i++ {
		cluster.ports[i] = basePort + i
	}

	// Build peer addresses for each node
	for i := 0; i < n; i++ {
		var peers []string
		for j := 0; j < n; j++ {
			if i != j {
				peers = append(peers, fmt.Sprintf("localhost:%d", cluster.ports[j]))
			}
		}

		// Create node
		cluster.nodes[i] = raft.NewRaftNode(i, peers)

		// Create gRPC server
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cluster.ports[i]))
		if err != nil {
			t.Fatalf("Failed to create listener for node %d: %v", i, err)
		}
		cluster.listeners[i] = listener

		server := grpc.NewServer()
		pb.RegisterRaftServiceServer(server, cluster.nodes[i])
		cluster.servers[i] = server

		// Start gRPC server
		go func(srv *grpc.Server, lis net.Listener) {
			srv.Serve(lis)
		}(server, listener)
	}

	return cluster
}

// Start starts all nodes in the cluster
func (tc *TestCluster) Start() {
	for i := 0; i < len(tc.nodes); i++ {
		tc.nodes[i].Start()
	}
}

// StartNode starts a specific node
func (tc *TestCluster) StartNode(id int) {
	if id >= 0 && id < len(tc.nodes) {
		tc.nodes[id].Start()
	}
}

// Stop stops all nodes and servers
func (tc *TestCluster) Stop() {
	for i := 0; i < len(tc.nodes); i++ {
		tc.nodes[i].Stop()
		tc.servers[i].Stop()
		tc.listeners[i].Close()
	}
}

// StopNode stops a specific node
func (tc *TestCluster) StopNode(id int) {
	if id >= 0 && id < len(tc.nodes) {
		tc.nodes[id].Stop()
	}
}

// GetLeader returns the ID of the current leader, or -1 if none
func (tc *TestCluster) GetLeader() int {
	for i := 0; i < len(tc.nodes); i++ {
		if tc.nodes[i].IsLeader() {
			return i
		}
	}
	return -1
}

// WaitForLeader waits for a leader to be elected within the timeout
func (tc *TestCluster) WaitForLeader(timeout time.Duration) int {
	start := time.Now()
	for time.Since(start) < timeout {
		leader := tc.GetLeader()
		if leader != -1 {
			return leader
		}
		time.Sleep(50 * time.Millisecond)
	}
	return -1
}

// CheckOneLeader verifies that exactly one leader exists
func (tc *TestCluster) CheckOneLeader() int {
	leaders := 0
	leaderID := -1

	for i := 0; i < len(tc.nodes); i++ {
		if tc.nodes[i].IsLeader() {
			leaders++
			leaderID = i
		}
	}

	if leaders != 1 {
		tc.t.Fatalf("Expected 1 leader, found %d", leaders)
	}

	return leaderID
}

// CheckNoLeader verifies that no leader exists
func (tc *TestCluster) CheckNoLeader() {
	for i := 0; i < len(tc.nodes); i++ {
		if tc.nodes[i].IsLeader() {
			tc.t.Fatalf("Expected no leader, but node %d is leader", i)
		}
	}
}

// GetNode returns a specific node
func (tc *TestCluster) GetNode(id int) *raft.RaftNode {
	if id >= 0 && id < len(tc.nodes) {
		return tc.nodes[id]
	}
	return nil
}

// CountNodes counts how many nodes are in a specific state
func (tc *TestCluster) CountNodesInState(state raft.NodeState) int {
	count := 0
	for i := 0; i < len(tc.nodes); i++ {
		if tc.nodes[i].GetState() == state {
			count++
		}
	}
	return count
}
