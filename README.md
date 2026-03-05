# Distributed KV-store w/ Raft Consensus

## Goal:
A kv database that can run on multiple computers where:
- If one computer dies, the others keep working
- All computers agree on what data they have (consensus)
- Clients can read/write data and it synchronizes


### 1. The Raft Protocol
Raft solves: "How do multiple computers agree on a sequence of operations?"
Three main pieces:
- Leader Election:

One node is the "leader" at any time
If leader dies, others elect a new one
Uses voting and terms (like election cycles)

- Log Replication:

Leader receives client requests
Leader adds them to its log
Leader sends log entries to followers
Once majority confirms, entry is "committed"

- Safety:

Ensures all nodes eventually have the same log in the same order
Even if networks partition or nodes crash

---

2. The Key-Value Store
This is just a map[string]string that gets updated by committed log entries.
When a log entry is committed, you apply it to your state machine (the map).

### High-Level Architecture

Components:

RaftNode
```
type RaftNode struct {
    // State
    currentTerm int
    votedFor    int
    log         []LogEntry
    
    // Role (follower, candidate, or leader)
    state       NodeState
    
    // The actual key-value data
    stateMachine map[string]string
    
    // Communication
    peers       []string  // addresses of other nodes
    
    // Timers
    electionTimer  *time.Timer
    heartbeatTimer *time.Timer
}
```

RPC Interface
```
// Nodes talk to each other via these RPCs
type RequestVote struct {
    Term         int
    CandidateId  int
    LastLogIndex int
    LastLogTerm  int
}

type AppendEntries struct {
    Term         int
    LeaderId     int
    Entries      []LogEntry
    LeaderCommit int
}
```

Client API
```
// Clients talk to cluster via these
type KVClient struct {
    // Maintains connections to all nodes
    // Automatically finds the leader
}

client.Put("key", "value")
value := client.Get("key")
```
