package raft

import (
	"context"
	"fmt"
	"time"

	pb "github.com/liambrem/dist-kv-store-raft/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (rn *RaftNode) sendRequestVote(peerAddr string, args *RequestVote, reply *RequestVoteReply) bool {
	// Create gRPC connection
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, err := grpc.Dial(peerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return false
	}
	defer conn.Close()

	client := pb.NewRaftServiceClient(conn)

	// Make RPC cal
	req := &pb.RequestVoteRequest{
		Term:         int32(args.Term),
		CandidateId:  int32(args.CandidateId),
		LastLogIndex: int32(args.LastLogIndex),
		LastLogTerm:  int32(args.LastLogTerm),
	}

	resp, err := client.RequestVote(ctx, req)
	if err != nil {
		return false
	}

	// Convert response
	reply.Term = int(resp.Term)
	reply.VoteGranted = resp.VoteGranted

	return true
}

// RequestVote implements the gRPC service
func (rn *RaftNode) RequestVote(ctx context.Context, req *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	resp := &pb.RequestVoteResponse{
		Term:        int32(rn.currentTerm),
		VoteGranted: false,
	}

	args := &RequestVote{
		Term:         int(req.Term),
		CandidateId:  int(req.CandidateId),
		LastLogIndex: int(req.LastLogIndex),
		LastLogTerm:  int(req.LastLogTerm),
	}

	// If candidate's term is outdated, reject
	if args.Term < rn.currentTerm {
		return resp, nil
	}

	// If candidate has higher term, update and step down
	if args.Term > rn.currentTerm {
		rn.currentTerm = args.Term
		rn.state = Follower
		rn.votedFor = -1
	}

	// Check if we can grant vote
	if rn.votedFor == -1 || rn.votedFor == args.CandidateId {
		lastLogIndex := len(rn.log) - 1
		lastLogTerm := 0
		if lastLogIndex >= 0 {
			lastLogTerm = rn.log[lastLogIndex].Term
		}

		logUpToDate := args.LastLogTerm > lastLogTerm ||
			(args.LastLogTerm == lastLogTerm && args.LastLogIndex >= lastLogIndex)

		if logUpToDate {
			rn.votedFor = args.CandidateId
			resp.VoteGranted = true

			// Reset election timeout
			select {
			case rn.heartbeatCh <- true:
			default:
			}
		}
	}

	resp.Term = int32(rn.currentTerm)
	return resp, nil
}

// AppendEntries implements the gRPC service for AppendEntries RPC
func (rn *RaftNode) AppendEntries(ctx context.Context, req *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	resp := &pb.AppendEntriesResponse{
		Term:    int32(rn.currentTerm),
		Success: false,
	}

	args := &AppendEntriesArgs{
		Term:         int(req.Term),
		LeaderId:     int(req.LeaderId),
		PrevLogIndex: int(req.PrevLogIndex),
		PrevLogTerm:  int(req.PrevLogTerm),
		LeaderCommit: int(req.LeaderCommit),
	}

	// Convert protobuf entries to LogEntry
	args.Entries = make([]LogEntry, len(req.Entries))
	for i, entry := range req.Entries {
		args.Entries[i] = LogEntry{
			Term:    int(entry.Term),
			Index:   int(entry.Index),
			Command: entry.Command,
		}
	}

	// Reply false if term < currentTerm
	if args.Term < rn.currentTerm {
		return resp, nil
	}

	// If term is higher or equal, this is a valid leader
	if args.Term >= rn.currentTerm {
		rn.currentTerm = args.Term
		rn.state = Follower
		rn.votedFor = -1

		// Reset election timeout by signaling heartbeat received
		select {
		case rn.heartbeatCh <- true:
			fmt.Printf("[Node %d] Received heartbeat from leader %d (term %d)\n", rn.id, args.LeaderId, args.Term)
		default:
		}
	}

	// For now, just accept heartbeats (empty entries)
	// Full log replication logic will be added later
	if len(args.Entries) == 0 {
		resp.Success = true
	}

	resp.Term = int32(rn.currentTerm)
	return resp, nil
}
