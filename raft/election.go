package raft

import (
	"time"
)

func (rn *RaftNode) runFollower() {
	timeout := randomElectionTimeout()
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-rn.heartbeatCh:
			// recieved heartbeat from leader
			timer.Reset(randomElectionTimeout())

		case <-timer.C:
			// timeout - become candidate
			rn.mu.Lock()
			rn.state = Candidate
			rn.mu.Unlock()
			return

		case <-rn.stopCh:
			return
		}
	}
}

func (rn *RaftNode) runCandidate() {
	rn.mu.Lock()
	rn.currentTerm++ // increment current term
	rn.votedFor = rn.id
	rn.mu.Unlock()

	timeout := randomElectionTimeout()
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	// start election
	go rn.requestVote()

	votesReceived := 1                 // vote for self
	votesNeeded := len(rn.peers)/2 + 1 // majority

	for {
		select {
		case <-rn.grantVoteCh:
			votesReceived++
			if votesReceived >= votesNeeded {
				// become leader
				rn.mu.Lock()
				rn.state = Leader
				rn.mu.Unlock()
				return
			}

		case <-rn.heartbeatCh:
			// another leader exists, step down
			rn.mu.Lock()
			rn.state = Follower
			rn.mu.Unlock()
			return

		case <-timer.C:
			// timeout - start new election
			return

		case <-rn.stopCh:
			return

		}
	}
}

func (rn *RaftNode) runLeader() {
	// Append Entries - implemented in rpc_handlers
	rn.mu.Lock()
	for i := range rn.nextIndex {
		rn.nextIndex[i] = len(rn.log)
	}
	for i := range rn.matchIndex {
		rn.matchIndex[i] = -1
	}
	rn.mu.Unlock()

	ticker := time.NewTicker(HeartbeatInterval)
	defer ticker.Stop()

	for {
		// Check if still leader (could have stepped down due to higher term)
		rn.mu.Lock()
		if rn.state != Leader {
			rn.mu.Unlock()
			return
		}
		rn.mu.Unlock()

		select {
		case <-ticker.C:
			go rn.sendHeartbeats()
		case <-rn.stopCh:
			return
		}
	}

}

func (rn *RaftNode) requestVote() {
	rn.mu.Lock()
	lastLogIndex := len(rn.log) - 1
	lastLogTerm := 0
	if lastLogIndex >= 0 {
		lastLogTerm = rn.log[lastLogIndex].Term
	}

	args := RequestVote{
		Term:         rn.currentTerm,
		CandidateId:  rn.id,
		LastLogIndex: lastLogIndex,
		LastLogTerm:  lastLogTerm,
	}
	rn.mu.Unlock()

	// Send RequestVote RPC to all peers in parallel
	for i, peer := range rn.peers {
		go func(peerAddr string, peerIndex int) {
			var reply RequestVoteReply

			// macke RPC call to peer
			ok := rn.sendRequestVote(peerAddr, &args, &reply)

			if !ok {
				return // RPC failed
			}

			rn.mu.Lock()
			defer rn.mu.Unlock()

			// if reply term is higher, step down
			if reply.Term > rn.currentTerm {
				rn.currentTerm = reply.Term
				rn.state = Follower
				rn.votedFor = -1
				return
			}

			// if vote granted and still a candidate
			if reply.VoteGranted && rn.state == Candidate {
				// signal vote received
				select {
				case rn.grantVoteCh <- true:
				default:
					// channel full, vote already counted
				}
			}
		}(peer, i)
	}
}
