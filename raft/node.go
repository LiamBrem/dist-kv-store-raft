package raft

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type RaftNode struct {
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
		stopCh:      make(chan struct{}),
	}
}

func (rn *RaftNode) Start() {
	go rn.run()
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
			rn.runFollower() // to implement

		case Candidate:
			fmt.Printf("[Node %d] Running as Candidate (term %d)\n", rn.id, rn.currentTerm)
			rn.runCandidate() // to implement

		case Leader:
			fmt.Printf("[Node %d] Running as Leader (term %d)\n", rn.id, rn.currentTerm)
			rn.runLeader() // to implement
		}

		// should stop
		select {
		case <-rn.stopCh:
			return
		default:
		}
	}
}
