package raft

import (
	"context"
	"time"

	pb "github.com/liambrem/dist-kv-store-raft/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// sendHeartbeats sends empty AppendEntries RPCs to all peers
func (rn *RaftNode) sendHeartbeats() {
	rn.mu.Lock()
	if rn.state != Leader {
		rn.mu.Unlock()
		return
	}

	args := AppendEntriesArgs{
		Term:         rn.currentTerm,
		LeaderId:     rn.id,
		Entries:      []LogEntry{}, // empty for heartbeat
		LeaderCommit: rn.commitIndex,
	}
	rn.mu.Unlock()

	// Send to all peers in parallel
	for _, peer := range rn.peers {
		go func(peerAddr string) {
			var reply AppendEntriesReply
			ok := rn.sendAppendEntries(peerAddr, &args, &reply)

			if !ok {
				return // RPC failed
			}

			rn.mu.Lock()
			defer rn.mu.Unlock()

			// If reply term is higher, step down
			if reply.Term > rn.currentTerm {
				rn.currentTerm = reply.Term
				rn.state = Follower
				rn.votedFor = -1
				return
			}
		}(peer)
	}
}

// sendAppendEntries makes the RPC call to a peer
func (rn *RaftNode) sendAppendEntries(peerAddr string, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, err := grpc.Dial(peerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return false
	}
	defer conn.Close()

	client := pb.NewRaftServiceClient(conn)

	// Convert LogEntry slice to protobuf
	pbEntries := make([]*pb.LogEntry, len(args.Entries))
	for i, entry := range args.Entries {
		pbEntries[i] = &pb.LogEntry{
			Term:    int32(entry.Term),
			Index:   int32(entry.Index),
			Command: entry.Command,
		}
	}

	req := &pb.AppendEntriesRequest{
		Term:         int32(args.Term),
		LeaderId:     int32(args.LeaderId),
		PrevLogIndex: int32(args.PrevLogIndex),
		PrevLogTerm:  int32(args.PrevLogTerm),
		Entries:      pbEntries,
		LeaderCommit: int32(args.LeaderCommit),
	}

	resp, err := client.AppendEntries(ctx, req)
	if err != nil {
		return false
	}

	// Convert response
	reply.Term = int(resp.Term)
	reply.Success = resp.Success

	return true
}
