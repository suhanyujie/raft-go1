package kv

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/raft"
)

type KvFsm struct {
	Db *sync.Map
}

type setPayload struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (kf *KvFsm) Apply(log *raft.Log) interface{} {
	switch log.Type {
	case raft.LogCommand:
		var sp setPayload
		err := json.Unmarshal(log.Data, &sp)
		if err != nil {
			return fmt.Errorf("Cound not parse payload: %s", err)
		}
		kf.Db.Store(sp.Key, sp.Value)
	default:
		return fmt.Errorf("Unknown raft log type: %#v", log.Type)
	}

	return nil
}

func (kf *KvFsm) Restore(rc io.ReadCloser) error {
	// Must always restore from a clean state!!
	kf.Db.Range(func(key interface{}, _ interface{}) bool {
		kf.Db.Delete(key)
		return true
	})

	decoder := json.NewDecoder(rc)

	for decoder.More() {
		var sp setPayload
		err := decoder.Decode(&sp)
		if err != nil {
			return fmt.Errorf("Could not decode payload: %s", err)
		}

		kf.Db.Store(sp.Key, sp.Value)
	}

	return rc.Close()
}

// Snapshot 快照功能暂不实现 todo
func (kf *KvFsm) Snapshot() (raft.FSMSnapshot, error) {
	return snapshotNoop{}, nil
}
