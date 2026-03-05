package tests

import (
	"testing"
	"time"
)

func TestLeaderFailure(t *testing.T) {
	cluster := NewTestCluster(t, 3)
	defer cluster.Stop()

	cluster.Start()

	leader1 := cluster.WaitForLeader(5 * time.Second)
	if leader1 == -1 {
		t.Fatal("No initial leader elected")
	}
	t.Logf("Initial leader: Node %d", leader1)

	cluster.StopNode(leader1)
	t.Logf("Killed leader Node %d", leader1)

	time.Sleep(1 * time.Second)

	leader2 := cluster.WaitForLeader(5 * time.Second)
	if leader2 == -1 {
		t.Fatal("No new leader elected after failure")
	}

	if leader2 == leader1 {
		t.Fatal("Same leader re-elected")
	}

	t.Logf("New leader elected: Node %d", leader2)
	cluster.CheckOneLeader()
}

func TestFollowerFailure(t *testing.T) {
	cluster := NewTestCluster(t, 5)
	defer cluster.Stop()

	cluster.Start()

	leaderID := cluster.WaitForLeader(5 * time.Second)
	if leaderID == -1 {
		t.Fatal("No leader elected")
	}

	followerID := -1
	for i := 0; i < 5; i++ {
		if i != leaderID {
			followerID = i
			break
		}
	}

	cluster.StopNode(followerID)
	t.Logf("Killed follower Node %d", followerID)

	time.Sleep(500 * time.Millisecond)

	currentLeader := cluster.CheckOneLeader()
	if currentLeader != leaderID {
		t.Fatalf("Leader changed after follower failure: was %d, now %d", leaderID, currentLeader)
	}

	t.Log("Leader remained stable after follower failure")
}

func TestMultipleFailures(t *testing.T) {
	cluster := NewTestCluster(t, 5)
	defer cluster.Stop()

	cluster.Start()

	leader1 := cluster.WaitForLeader(5 * time.Second)
	if leader1 == -1 {
		t.Fatal("No initial leader")
	}
	t.Logf("Leader 1: Node %d", leader1)

	cluster.StopNode(leader1)
	time.Sleep(1 * time.Second)

	leader2 := cluster.WaitForLeader(5 * time.Second)
	if leader2 == -1 || leader2 == leader1 {
		t.Fatal("Second leader not elected")
	}
	t.Logf("Leader 2: Node %d", leader2)

	for i := 0; i < 5; i++ {
		if i != leader1 && i != leader2 {
			cluster.StopNode(i)
			t.Logf("Killed Node %d", i)
			break
		}
	}

	time.Sleep(500 * time.Millisecond)

	cluster.CheckOneLeader()
	t.Log("Cluster maintains quorum with 3 nodes")
}

func TestQuorumFailure(t *testing.T) {
	cluster := NewTestCluster(t, 5)
	defer cluster.Stop()

	cluster.Start()

	leader := cluster.WaitForLeader(5 * time.Second)
	if leader == -1 {
		t.Fatal("No initial leader")
	}

	killed := 0
	for i := 0; i < 5 && killed < 3; i++ {
		cluster.StopNode(i)
		killed++
		t.Logf("Killed Node %d", i)
	}

	time.Sleep(2 * time.Second)

	cluster.CheckNoLeader()
	t.Log("Correctly no leader without quorum")
}

func TestRecoveryAfterFailure(t *testing.T) {
	t.Skip("Skipping recovery test - requires restart capability")
}

func TestRapidLeaderChanges(t *testing.T) {
	cluster := NewTestCluster(t, 5)
	defer cluster.Stop()

	cluster.Start()

	leaders := make([]int, 0)

	for i := 0; i < 3; i++ {
		leader := cluster.WaitForLeader(5 * time.Second)
		if leader == -1 {
			t.Fatalf("No leader elected in iteration %d", i)
		}

		leaders = append(leaders, leader)
		t.Logf("Iteration %d: Leader is Node %d", i, leader)

		cluster.StopNode(leader)
		time.Sleep(1 * time.Second)
	}

	finalLeader := cluster.WaitForLeader(5 * time.Second)
	if finalLeader == -1 {
		t.Fatal("No final leader elected")
	}

	t.Logf("Final leader: Node %d", finalLeader)
	cluster.CheckOneLeader()
}
