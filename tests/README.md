# Raft Tests

This directory contains comprehensive tests for the Raft consensus implementation.

## Test Categories

### 1. Election Tests ([election_test.go](election_test.go))
Tests for leader election mechanism:
- `TestInitialElection` - Verifies a leader is elected on cluster start
- `TestReElection` - Tests new leader election after failure
- `TestElectionWithFiveNodes` - Tests election in larger cluster
- `TestMultipleElections` - Tests repeated leader failures
- `TestNoElectionWithoutQuorum` - Verifies no leader without majority

### 2. Replication Tests ([replication_test.go](replication_test.go))
Tests for log replication and command proposals:
- `TestBasicProposal` - Leader accepts command proposals
- `TestMultipleProposals` - Multiple sequential commands
- `TestFollowerRejectsProposal` - Followers reject proposals
- `TestProposalAfterLeaderChange` - Commands after leader change
- `TestConcurrentProposals` - Concurrent command submission

### 3. Failure Tests ([failure_test.go](failure_test.go))
Tests for fault tolerance:
- `TestLeaderFailure` - Recovery from leader crash
- `TestFollowerFailure` - Cluster continues with follower failure
- `TestMultipleFailures` - Multiple node failures
- `TestQuorumFailure` - No leader without quorum
- `TestRapidLeaderChanges` - Stability under rapid failures

### 4. Partition Tests ([partition_test.go](partition_test.go))
Network partition scenarios (infrastructure needed):
- `TestMajorityPartition` - Majority/minority split
- `TestMinorityPartition` - Minority behavior
- `TestPartitionWithLeaderInMinority` - Leader in minority
- `TestPartitionHealing` - Partition recovery
- `TestSymmetricPartition` - Equal split scenario

## Running Tests

### Run All Tests
```bash
cd tests
go test -v
```

### Run Specific Test Category
```bash
# Election tests only
go test -v -run Election

# Failure tests only
go test -v -run Failure

# Replication tests only
go test -v -run Proposal
```

### Run Individual Test
```bash
go test -v -run TestInitialElection
```

### Run with Timeout
```bash
go test -v -timeout 30s
```

## Test Helpers ([helpers.go](helpers.go))

The test helpers provide utilities for managing test clusters:

- `NewTestCluster(t, n)` - Creates n-node cluster
- `cluster.Start()` - Starts all nodes
- `cluster.Stop()` - Stops all nodes
- `cluster.WaitForLeader(timeout)` - Waits for leader election
- `cluster.CheckOneLeader()` - Verifies exactly one leader
- `cluster.GetNode(id)` - Gets specific node
- `cluster.StopNode(id)` - Stops specific node

## Expected Results

### Passing Tests
Currently, these tests should pass:
- All election tests
- Basic proposal tests
- Most failure tests (except those requiring node restart)

### Skipped Tests
These tests are skipped pending additional features:
- Partition tests (require network simulation infrastructure)
- Recovery tests (require node restart capability)

## Notes

1. **Ports**: Tests use ports 10000+ to avoid conflicts
2. **Timeouts**: Election timeout is 150-300ms, test timeouts are generous
3. **Concurrency**: Tests run with Go's race detector enabled via `-race` flag
4. **Cleanup**: All tests properly cleanup resources via `defer cluster.Stop()`

## Future Enhancements

To enable full partition testing:
1. Implement network proxy layer for selective packet dropping
2. Add cluster partition/heal methods
3. Implement node restart capability
4. Add log consistency verification
5. Add snapshot testing when implemented
