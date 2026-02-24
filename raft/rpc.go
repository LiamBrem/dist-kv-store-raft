package raft

// RequestVote is sent by candidates during elections
type RequestVote struct {
	Term int // current term of candidate
	CandidateId int // who's requesting vote
	LastLogIndex int // index of candidate's last log entry
	LastLogTerm int // term of candidates last log entry
}

type RequestVoteReply struct {
	Term int // responder's current term
	VoteGranted bool // whether vote was given
}

// AppendEntries is sent by leader for log replication/heartbeats
type AppendEntriesArgs struct {
	Term int // leaders current term
	LeaderId int
	PrevLogIndex int
	PrevLogTerm int
	Entries []LogEntry // empty for heartbeat
	LeaderCommit int // leader's commit index
}

type AppendEntriesReply struct {
	Term int // Responder's current term
	Success bool // if follower accepted entries
}