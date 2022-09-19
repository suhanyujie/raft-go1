package kv

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/raft"
)

type KvFsm struct {
	db *sync.Map
}

type setPayload struct {
	Key, Value string
}

func (kf *KvFsm) apply(log *raft.Log) error {
	switch log.Type {
	case raft.LogCommand:
		var sp setPayload
		err := json.Unmarshal(log.Data, &sp)
		if err != nil {
			return fmt.Errorf("Cound not parse payload: %s", err)
		}
		kf.db.Store(sp.Key, sp.Value)
	default:
		return fmt.Errorf("Unknown raft log type: %#v", log.Type)
	}

	return nil
}

func (kf *KvFsm) Restore(rc io.ReadCloser) error {
	// Must always restore from a clean state!!
	kf.db.Range(func(key interface{}, _ interface{}) bool {
		kf.db.Delete(key)
		return true
	})

	decoder := json.NewDecoder(rc)

	for decoder.More() {
		var sp setPayload
		err := decoder.Decode(&sp)
		if err != nil {
			return fmt.Errorf("Could not decode payload: %s", err)
		}

		kf.db.Store(sp.Key, sp.Value)
	}

	return rc.Close()
}

func (kf *KvFsm) Snapshot() (raft.FSMSnapshot, error) {
	return snapshotNoop{}, nil
}
