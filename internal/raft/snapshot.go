package raft_internal

import (
	"encoding/json"
	"kv-storage/internal/store"

	"github.com/hashicorp/raft"
)

type fsmSnapshot struct {
	store *store.Store
}

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	defer sink.Close()

	data, err := f.store.GetAll()
	if err != nil {
		sink.Cancel()
		return err
	}

	if err := json.NewEncoder(sink).Encode(data); err != nil {
		sink.Cancel()
		return err
	}

	return nil
}

func (f *fsmSnapshot) Release() {
}
