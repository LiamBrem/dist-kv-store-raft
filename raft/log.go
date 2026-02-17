package raft

// Represents a single entry in raft log

type LogEntry struct {
	Term int // which term entry was created in
	Index int // position in log
	Command string // e.g. "PUT ..."
}