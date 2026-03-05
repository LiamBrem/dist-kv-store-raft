package tests

import (
	"testing"
	"time"
)

// TestBasicProposal tests that a leader can accept command proposals
func TestBasicProposal(t *testing.T) {
	cluster := NewTestCluster(t, 3)
	defer cluster.Stop()

	cluster.Start()

	// Wait for leader
	leaderID := cluster.WaitForLeader(5 * time.Second)
	if leaderID == -1 {
		t.Fatal("No leader elected")
	}

	leader := cluster.GetNode(leaderID)

	// Propose a command
	index, term, isLeader := leader.ProposeCommand("PUT x 100")
	if !isLeader {
		t.Fatal("Leader did not accept command")
	}

	if index < 0 || term < 1 {
		t.Fatalf("Invalid index=%d or term=%d", index, term)
	}

	t.Logf("Command proposed at index=%d, term=%d", index, term)
}

// TestMultipleProposals tests that a leader can accept multiple commands
func TestMultipleProposals(t *testing.T) {
	cluster := NewTestCluster(t, 3)
	defer cluster.Stop()

	cluster.Start()

	leaderID := cluster.WaitForLeader(5 * time.Second)
	if leaderID == -1 {
		t.Fatal("No leader elected")
	}

	leader := cluster.GetNode(leaderID)

	// Propose multiple commands
	commands := []string{
		"PUT x 1",
		"PUT y 2",
		"PUT z 3",
	}

	for i, cmd := range commands {
		index, term, isLeader := leader.ProposeCommand(cmd)
		if !isLeader {
			t.Fatalf("Leader rejected command %d", i)
		}
		if index != i {
			t.Fatalf("Expected index %d, got %d", i, index)
		}
		t.Logf("Command %d: index=%d, term=%d", i, index, term)
	}
}

// TestFollowerRejectsProposal tests that followers reject command proposals
func TestFollowerRejectsProposal(t *testing.T) {
	cluster := NewTestCluster(t, 3)
	defer cluster.Stop()

	cluster.Start()

	leaderID := cluster.WaitForLeader(5 * time.Second)
	if leaderID == -1 {
		t.Fatal("No leader elected")
	}

	// Find a follower
	var followerID int
	for i := 0; i < 3; i++ {
		if i != leaderID {
			followerID = i
			break
		}
	}

	follower := cluster.GetNode(followerID)

	// Try to propose to follower
	_, _, isLeader := follower.ProposeCommand("PUT x 100")
	if isLeader {
		t.Fatal("Follower incorrectly accepted command proposal")
	}

	t.Logf("Follower correctly rejected proposal")
}

// TestProposalAfterLeaderChange tests proposals after a leader change
func TestProposalAfterLeaderChange(t *testing.T) {
	cluster := NewTestCluster(t, 3)
	defer cluster.Stop()

	cluster.Start()

	// Get initial leader
	leader1ID := cluster.WaitForLeader(5 * time.Second)
	if leader1ID == -1 {
		t.Fatal("No initial leader elected")
	}

	leader1 := cluster.GetNode(leader1ID)
	index1, _, _ := leader1.ProposeCommand("PUT x 1")
	t.Logf("First leader (Node %d) proposed command at index %d", leader1ID, index1)

	// Stop the leader
	cluster.StopNode(leader1ID)
	time.Sleep(1 * time.Second)

	// Wait for new leader
	leader2ID := cluster.WaitForLeader(5 * time.Second)
	if leader2ID == -1 || leader2ID == leader1ID {
		t.Fatal("New leader not elected")
	}

	leader2 := cluster.GetNode(leader2ID)
	time.Sleep(500 * time.Millisecond) // Let new leader stabilize

	// Propose to new leader
	index2, _, isLeader := leader2.ProposeCommand("PUT y 2")
	if !isLeader {
		t.Fatal("New leader rejected command")
	}

	t.Logf("New leader (Node %d) proposed command at index %d", leader2ID, index2)
}

// TestConcurrentProposals tests multiple concurrent proposals
func TestConcurrentProposals(t *testing.T) {
	cluster := NewTestCluster(t, 3)
	defer cluster.Stop()

	cluster.Start()

	leaderID := cluster.WaitForLeader(5 * time.Second)
	if leaderID == -1 {
		t.Fatal("No leader elected")
	}

	leader := cluster.GetNode(leaderID)

	// Propose commands concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			index, _, isLeader := leader.ProposeCommand("PUT key value")
			if !isLeader {
				t.Errorf("Command %d rejected", id)
			}
			t.Logf("Concurrent command %d at index %d", id, index)
			done <- true
		}(i)
	}

	// Wait for all proposals
	for i := 0; i < 10; i++ {
		<-done
	}

	t.Log("All concurrent proposals completed")
}
