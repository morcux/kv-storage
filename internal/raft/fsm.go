package raft_internal

import (
	"encoding/json"
	"io"
	"kv-storage/internal/store"

	"github.com/hashicorp/raft"
)

type FSM struct {
	Store *store.Store
}

type Command struct {
	Op    string // "SET", "DELETE"
	Key   string
	Value []byte
}

func (f *FSM) Apply(log *raft.Log) interface{} {
	var cmd Command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		return err
	}

	switch cmd.Op {
	case "SET":
		return f.Store.Set(cmd.Key, cmd.Value)
	}
	return nil
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return &fsmSnapshot{store: f.Store}, nil
}

func (f *FSM) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	newData := make(map[string][]byte)
	if err := json.NewDecoder(rc).Decode(&newData); err != nil {
		return err
	}

	return f.Store.RestoreFromMap(newData)
}
