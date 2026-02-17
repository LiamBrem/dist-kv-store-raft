# Notes

For node/server implementation, build one program that acts as a raft node and then run X (for example 5) copies of it locally (each listening on a different port).

They'll be communicating through RPC

Each node:
- Runs as a separate process
- Listens on its own port for incoming RPCs
- Has a list of all other nodes' addresses
- Can send RPCs to any other node


### **For Automated Testing:**

```go
// test/cluster_test.go
func TestLeaderElection(t *testing.T) {
    // Start 3 nodes programmatically
    nodes := make([]*RaftNode, 3)
    for i := 0; i < 3; i++ {
        nodes[i] = NewRaftNode(i, 8000+i, getPeerAddresses(i, 3))
        nodes[i].Start()
    }
    
    time.Sleep(5 * time.Second)
    
    // Check that exactly one is leader
    leaders := 0
    for _, node := range nodes {
        if node.state == Leader {
            leaders++
        }
    }
    
    if leaders != 1 {
        t.Errorf("Expected 1 leader, got %d", leaders)
    }
}
```

## **Network Partitions & Failure Simulation**

Once you have basic functionality working, you can simulate failures:

```go
// Simulate node crash
func (rn *RaftNode) Crash() {
    rn.crashed = true
    // Stop responding to RPCs
}

// Simulate network partition
func (rn *RaftNode) PartitionFrom(nodeIds []int) {
    // Stop responding to RPCs from these nodes
}
```


Test examples:
- TestLeaderElection - start 5 nodes, verify one becomes leader
- TestLeaderFailure - kill the leader, verify new one elected
- TestNetworkPartition - split cluster 3-2, verify majority side still works
- TestLogReplication - send commands, verify all nodes have same log
- TestConcurrentWrites - multiple clients writing simultaneously
- TestNodeRecovery - crash a node, restart it, verify it catches up
- TestMinorityFailure - kill 2 of 5 nodes, verify cluster still functions