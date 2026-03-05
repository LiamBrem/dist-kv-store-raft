package raft

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	pb "github.com/liambrem/dist-kv-store-raft/proto"
)

// ApplyMsg is sent to the application when a log entry is committed
type ApplyMsg struct {
	CommandValid bool
	Command      string
	CommandIndex int
}

type RaftNode struct {
	pb.UnimplementedRaftServiceServer
	mu sync.Mutex

	id    int
	peers []string // addresses of other nodes ["localhost:8003"]

	currentTerm int
	votedFor    int // -1: no vote this term
	log         []LogEntry

	state       NodeState
	commitIndex int // highest log index known to be committed
	lastApplied int // highest log index applied to state machine

	// Leader only
	nextIndex  []int // next log index to send for each peer
	matchIndex []int // highest log index known replicated for each peer

	// for internal commmunication
	heartbeatCh chan bool     // when heartbeat arrives
	grantVoteCh chan bool     // when vote is granted
	leaderCh    chan bool     // when node becomes leader
	commitCh    chan bool     // when new entries committed
	applyCh     chan ApplyMsg // send committed commands to application
	stopCh      chan struct{} // to shut down
}

func NewRaftNode(id int, peers []string) *RaftNode {
	return &RaftNode{
		id:          id,
		peers:       peers,
		state:       Follower,
		currentTerm: 0,
		votedFor:    -1,
		log:         []LogEntry{},
		commitIndex: 0,
		lastApplied: 0,
		nextIndex:   make([]int, len(peers)),
		matchIndex:  make([]int, len(peers)),
		heartbeatCh: make(chan bool, 1),
		grantVoteCh: make(chan bool, 1),
		leaderCh:    make(chan bool, 1),
		commitCh:    make(chan bool, 1),
		applyCh:     make(chan ApplyMsg, 100),
		stopCh:      make(chan struct{}),
	}
}

func (rn *RaftNode) Start() {
	go rn.run()
	go rn.applyCommitted() // Start apply loop
}

func (rn *RaftNode) Stop() {
	close(rn.stopCh)
}

// returns random timeout between min and max
func randomElectionTimeout() time.Duration {
	diff := ElectionTimeoutMax - ElectionTimeoutMin
	return ElectionTimeoutMin + time.Duration(rand.Int63n(int64(diff)))
}

// main loop

func (rn *RaftNode) run() {
	for {
		rn.mu.Lock()
		state := rn.state
		rn.mu.Unlock()

		switch state {
		case Follower:
			fmt.Printf("[Node %d] Running as Follower (term %d)\n", rn.id, rn.currentTerm)
			rn.runFollower()

		case Candidate:
			fmt.Printf("[Node %d] Running as Candidate (term %d)\n", rn.id, rn.currentTerm)
			rn.runCandidate()

		case Leader:
			fmt.Printf("[Node %d] Running as Leader (term %d)\n", rn.id, rn.currentTerm)
			rn.runLeader()
		}

		// should stop
		select {
		case <-rn.stopCh:
			return
		default:
		}
	}
}

// GetApplyCh returns the channel for committed commands
func (rn *RaftNode) GetApplyCh() chan ApplyMsg {
	return rn.applyCh
}

// ProposeCommand submits a new command to the Raft log
// Returns: (index, term, isLeader)
func (rn *RaftNode) ProposeCommand(command string) (int, int, bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if rn.state != Leader {
		return -1, -1, false
	}

	// Create new log entry
	index := len(rn.log)
	entry := LogEntry{
		Term:    rn.currentTerm,
		Index:   index,
		Command: command,
	}

	// Append to log
	rn.log = append(rn.log, entry)
	fmt.Printf("[Node %d] Appended command to log at index %d: %s\n", rn.id, index, command)

	// TODO: Trigger replication to followers
	// For now, since we don't have full replication, we'll just commit locally
	// In a full implementation, we'd wait for majority replication

	return index, rn.currentTerm, true
}

// applyCommitted watches for newly committed log entries and applies them
func (rn *RaftNode) applyCommitted() {
	for {
		select {
		case <-rn.stopCh:
			return
		default:
		}

		rn.mu.Lock()
		// Apply any entries between lastApplied and commitIndex
		for rn.lastApplied < rn.commitIndex {
			rn.lastApplied++
			if rn.lastApplied < len(rn.log) {
				entry := rn.log[rn.lastApplied]
				msg := ApplyMsg{
					CommandValid: true,
					Command:      entry.Command,
					CommandIndex: rn.lastApplied,
				}

				rn.mu.Unlock()
				rn.applyCh <- msg
				rn.mu.Lock()
			}
		}
		rn.mu.Unlock()

		time.Sleep(10 * time.Millisecond)
	}
}
