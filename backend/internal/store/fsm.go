package store

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/hashicorp/raft"
	"github.com/hemu1808/auradeploy/backend/internal/models"
)

// ensure interface is implemented
var _ raft.FSM = &fsm{}

type fsm struct {
	mu   sync.RWMutex
	apps map[string]models.Application
}

func newFSM() *fsm {
	return &fsm{
		apps: make(map[string]models.Application),
	}
}

// Commands
type commandType string

const (
	SetAppCommand    commandType = "SET_APP"
	DeleteAppCommand commandType = "DELETE_APP"
)

type command struct {
	Type commandType        `json:"type"`
	App  models.Application `json:"app,omitempty"`
	ID   string             `json:"id,omitempty"`
}

// Apply applies a Raft log entry to the key-value store.
func (f *fsm) Apply(l *raft.Log) interface{} {
	var c command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		return err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	switch c.Type {
	case SetAppCommand:
		f.apps[c.App.ID] = c.App
	case DeleteAppCommand:
		delete(f.apps, c.ID)
	}

	return nil
}

// Snapshot returns a snapshot of the key-value store.
func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Clone the map
	clone := make(map[string]models.Application)
	for k, v := range f.apps {
		clone[k] = v
	}

	return &fsmSnapshot{apps: clone}, nil
}

// Restore stores the key-value store to a previous state.
func (f *fsm) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	var apps map[string]models.Application
	if err := json.NewDecoder(rc).Decode(&apps); err != nil {
		return err
	}

	f.mu.Lock()
	f.apps = apps
	f.mu.Unlock()

	return nil
}

type fsmSnapshot struct {
	apps map[string]models.Application
}

func (s *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		b, err := json.Marshal(s.apps)
		if err != nil {
			return err
		}

		if _, err := sink.Write(b); err != nil {
			return err
		}

		return sink.Close()
	}()

	if err != nil {
		sink.Cancel()
	}

	return err
}

func (s *fsmSnapshot) Release() {}
