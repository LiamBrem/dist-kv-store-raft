package kv

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/liambrem/dist-kv-store-raft/raft"
)

// Command types
const (
	OpPut    = "PUT"
	OpDelete = "DELETE"
)

type Command struct {
	Op    string // "PUT" or "DELETE"
	Key   string
	Value string
}

type KVStore struct {
	mu   sync.RWMutex
	data map[string]string // the actual key-value store
	raft *raft.RaftNode

	applyCh chan raft.ApplyMsg // receives committed commands from Raft
	stopCh  chan struct{}
}

// Creates new
func NewKVStore(raftNode *raft.RaftNode) *KVStore {
	kv := &KVStore{
		data:    make(map[string]string),
		raft:    raftNode,
		applyCh: raftNode.GetApplyCh(),
		stopCh:  make(chan struct{}),
	}

	go kv.applyLoop()

	return kv
}

// Get retrieves a value for a key (read-only, no consensus needed)
func (kv *KVStore) Get(key string) (string, bool) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	value, exists := kv.data[key]
	return value, exists
}

// Put sets a key-value pair (requires Raft consensus)
func (kv *KVStore) Put(key, value string) error {
	cmd := Command{
		Op:    OpPut,
		Key:   key,
		Value: value,
	}

	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %v", err)
	}

	index, term, isLeader := kv.raft.ProposeCommand(string(cmdBytes))
	if !isLeader {
		return errors.New("not the leader")
	}

	fmt.Printf("[KVStore] Proposed PUT %s=%s at index %d, term %d\n", key, value, index, term)

	time.Sleep(500 * time.Millisecond)

	return nil
}

// Delete removes a key (requires Raft consensus)
func (kv *KVStore) Delete(key string) error {
	cmd := Command{
		Op:  OpDelete,
		Key: key,
	}

	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %v", err)
	}

	index, term, isLeader := kv.raft.ProposeCommand(string(cmdBytes))
	if !isLeader {
		return errors.New("not the leader")
	}

	fmt.Printf("[KVStore] Proposed DELETE %s at index %d, term %d\n", key, index, term)

	time.Sleep(500 * time.Millisecond)

	return nil
}

// applyLoop processes committed commands from Raft
func (kv *KVStore) applyLoop() {
	for {
		select {
		case msg := <-kv.applyCh:
			if msg.CommandValid {
				kv.applyCommand(msg.Command)
			}
		case <-kv.stopCh:
			return
		}
	}
}

// applyCommand applies a committed command to the state machine
func (kv *KVStore) applyCommand(cmdStr string) {
	var cmd Command
	if err := json.Unmarshal([]byte(cmdStr), &cmd); err != nil {
		fmt.Printf("[KVStore] Failed to unmarshal command: %v\n", err)
		return
	}

	kv.mu.Lock()
	defer kv.mu.Unlock()

	switch cmd.Op {
	case OpPut:
		kv.data[cmd.Key] = cmd.Value
		fmt.Printf("[KVStore] Applied: PUT %s=%s (total keys: %d)\n", cmd.Key, cmd.Value, len(kv.data))
	case OpDelete:
		delete(kv.data, cmd.Key)
		fmt.Printf("[KVStore] Applied: DELETE %s (total keys: %d)\n", cmd.Key, len(kv.data))
	default:
		fmt.Printf("[KVStore] Unknown operation: %s\n", cmd.Op)
	}
}

func (kv *KVStore) GetAllKeys() []string {
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	keys := make([]string, 0, len(kv.data))
	for k := range kv.data {
		keys = append(keys, k)
	}
	return keys
}

func (kv *KVStore) Stop() {
	close(kv.stopCh)
}
