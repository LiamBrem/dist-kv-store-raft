package raft

import "time"

type NodeState int 

const (
	Follower NodeState = iota
	Candidate NodeState = iota
	Leader NodeState = iota
)

const (
	HeartbeatInterval = 100 * time.Millisecond
	ElectiontimeoutMin = 150 * time.Millisecond
	ElectionTimeoutMax = 300 * time.Millisecond
)