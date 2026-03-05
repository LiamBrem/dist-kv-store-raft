package tests

import (
	"testing"
	"time"
)

func TestInitialElection(t *testing.T) {
	cluster := NewTestCluster(t, 3)
	defer cluster.Stop()

	cluster.Start()

	leader := cluster.WaitForLeader(5 * time.Second)
	if leader == -1 {
		t.Fatal("No leader elected within timeout")
	}

	t.Logf("Leader elected: Node %d", leader)

	cluster.CheckOneLeader()
}

func TestReElection(t *testing.T) {
	cluster := NewTestCluster(t, 3)
	defer cluster.Stop()

	cluster.Start()

	leader1 := cluster.WaitForLeader(5 * time.Second)
	if leader1 == -1 {
		t.Fatal("No initial leader elected")
	}

	t.Logf("Initial leader: Node %d", leader1)

	cluster.StopNode(leader1)
	t.Logf("Stopped leader node %d", leader1)

	time.Sleep(1 * time.Second) // Give time for election timeout
	leader2 := cluster.WaitForLeader(5 * time.Second)
	if leader2 == -1 {
		t.Fatal("No new leader elected after old leader failed")
	}

	if leader2 == leader1 {
		t.Fatal("New leader is the same as old leader")
	}

	t.Logf("New leader elected: Node %d", leader2)
	cluster.CheckOneLeader()
}

func TestElectionWithFiveNodes(t *testing.T) {
	cluster := NewTestCluster(t, 5)
	defer cluster.Stop()

	cluster.Start()

	leader := cluster.WaitForLeader(5 * time.Second)
	if leader == -1 {
		t.Fatal("No leader elected in 5-node cluster")
	}

	t.Logf("Leader elected in 5-node cluster: Node %d", leader)
	cluster.CheckOneLeader()
}

func TestMultipleElections(t *testing.T) {
	cluster := NewTestCluster(t, 5)
	defer cluster.Stop()

	cluster.Start()

	leader1 := cluster.WaitForLeader(5 * time.Second)
	if leader1 == -1 {
		t.Fatal("No initial leader elected")
	}
	t.Logf("First leader: Node %d", leader1)

	cluster.StopNode(leader1)
	time.Sleep(1 * time.Second)

	leader2 := cluster.WaitForLeader(5 * time.Second)
	if leader2 == -1 || leader2 == leader1 {
		t.Fatal("Second leader election failed")
	}
	t.Logf("Second leader: Node %d", leader2)

	cluster.StopNode(leader2)
	time.Sleep(1 * time.Second)

	leader3 := cluster.WaitForLeader(5 * time.Second)
	if leader3 == -1 || leader3 == leader1 || leader3 == leader2 {
		t.Fatal("Third leader election failed")
	}
	t.Logf("Third leader: Node %d", leader3)

	cluster.CheckOneLeader()
}

func TestNoElectionWithoutQuorum(t *testing.T) {
	cluster := NewTestCluster(t, 5)
	defer cluster.Stop()

	cluster.Start()

	leader := cluster.WaitForLeader(5 * time.Second)
	if leader == -1 {
		t.Fatal("No initial leader elected")
	}

	cluster.StopNode(leader)
	for i := 0; i < 5; i++ {
		if i != leader {
			cluster.StopNode(i)
			break
		}
	}
	for i := 0; i < 5; i++ {
		if i != leader {
			cluster.StopNode(i)
			break
		}
	}

	time.Sleep(2 * time.Second)

	cluster.CheckNoLeader()
	t.Log("Correctly no leader elected without quorum")
}
