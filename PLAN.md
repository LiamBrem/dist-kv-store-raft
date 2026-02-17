## **How Raft Actually Works (Simplified)**

### **Normal Operation:**

1. **Client sends PUT("name", "Alice") to any node**
2. **That node forwards to leader** (if it's not the leader)
3. **Leader adds entry to its log**
4. **Leader sends AppendEntries RPC to all followers**
5. **Followers add entry to their logs and respond**
6. **Once majority responds, leader commits the entry**
7. **Leader applies to state machine** (updates the map)
8. **Leader tells followers to commit**
9. **Client gets success response**

### **When Leader Dies:**

1. **Followers stop receiving heartbeats**
2. **Election timers fire**
3. **Some follower becomes candidate**
4. **Candidate requests votes from others**
5. **If it gets majority, becomes new leader**
6. **New leader starts sending heartbeats**

## **Step-by-Step Build Plan**

### Part 1 

**Resources**
- [Raft paper](https://raft.github.io/raft.pdf)
- [Raft visualization](http://thesecretlivesofdata.com/raft/) 
- MIT 6.824 Lab 2 description

### Part 2 

**Goal:** Get nodes to elect a leader

**What to build:**
```go
// Start as follower
// If election timeout expires → become candidate
// Request votes from peers
// If get majority → become leader
// Send heartbeats as leader
```

**How to test:**
- Start 3 nodes
- Verify one becomes leader
- Kill leader, verify new one elected
- Print lots of logs to see what's happening

**Milestones:**
- Basic node structure, timers
- RequestVote RPC, vote counting
- Leader sends heartbeats, election works

### Part 3 

**Goal:** Leader can replicate log entries to followers

**What to build:**
```go
// Leader receives command (from client or test)
// Leader appends to its log
// Leader sends AppendEntries to followers
// Followers append and respond
// Leader waits for majority
// Leader commits and applies to state machine
```

**How to test:**
- Send commands to leader
- Verify all nodes have same log
- Kill some followers, verify still works with majority

**Milestones:**
- AppendEntries RPC, log structure
- Matching log entries, handling conflicts
- Commit index tracking, applying to state machine

### Part 4

**Goal:** GET/PUT operations work

**What to build:**
```go
type KVStore struct {
    raft *RaftNode
    data map[string]string
}

func (kv *KVStore) Put(key, value string) error {
    // Submit to Raft as a command
    // Wait for it to be committed
    // Return success
}
```

**Milestones:**
- State machine (the map), apply logic
- Client operations, leader forwarding

### Part 5

**Goal:** Easy-to-use client, handle edge cases

**What to build:**
- Client automatically finds leader
- Retries on failure
- Handle leader changes mid-request
- Better error handling

### Part 5

**Benchmarks + tests**
- Network partition tests
- Node failure/recovery tests
- Performance benchmarks
- Measure throughput and latency

