package store

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/hemu1808/auradeploy/backend/internal/models"
)

type Store struct {
	raft *raft.Raft
	fsm  *fsm
}

// NewStore initializes a new Raft store node
func NewStore(nodeID string, bindAddr string, raftDir string, bootstrap bool) (*Store, error) {
	// Setup Raft configuration
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(nodeID)

	// Setup Raft communication
	addr, err := net.ResolveTCPAddr("tcp", bindAddr)
	if err != nil {
		return nil, err
	}

	transport, err := raft.NewTCPTransport(bindAddr, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, err
	}

	// Create peer storage (BoltDB)
	err = os.MkdirAll(raftDir, 0700)
	if err != nil {
		return nil, err
	}

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(raftDir, "logs.dat"))
	if err != nil {
		return nil, err
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(raftDir, "stable.dat"))
	if err != nil {
		return nil, err
	}

	snapshotStore, err := raft.NewFileSnapshotStore(raftDir, 3, os.Stderr)
	if err != nil {
		return nil, err
	}

	// Instantiate the state machine
	fsm := newFSM()

	// Instantiate the Raft system
	r, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return nil, err
	}

	s := &Store{
		raft: r,
		fsm:  fsm,
	}

	if bootstrap {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}

		r.BootstrapCluster(configuration)
	}

	return s, nil
}

// GetApplications returns a copy of all applications in the FSM state
func (s *Store) GetApplications() map[string]models.Application {
	s.fsm.mu.RLock()
	defer s.fsm.mu.RUnlock()

	clone := make(map[string]models.Application)
	for k, v := range s.fsm.apps {
		clone[k] = v
	}

	return clone
}

// SetApplication proposes a state change to Raft: creating or replacing an Application
func (s *Store) SetApplication(app models.Application) error {
	if s.raft.State() != raft.Leader {
		return fmt.Errorf("not the leader")
	}

	c := command{
		Type: SetAppCommand,
		App:  app,
	}

	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	future := s.raft.Apply(b, 5*time.Second)
	if err := future.Error(); err != nil {
		return err
	}

	return nil
}

// DeleteApplication proposes a state change to Raft: removing an Application
func (s *Store) DeleteApplication(appID string) error {
	if s.raft.State() != raft.Leader {
		return fmt.Errorf("not the leader")
	}

	c := command{
		Type: DeleteAppCommand,
		ID:   appID,
	}

	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	future := s.raft.Apply(b, 5*time.Second)
	if err := future.Error(); err != nil {
		return err
	}

	return nil
}

// Join adds a new node to the cluster
func (s *Store) Join(nodeID, bindAddr string) error {
	if s.raft.State() != raft.Leader {
		return fmt.Errorf("not the leader")
	}

	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed and added again.
		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(bindAddr) {
			
			// However if both match, then we already joined, ignore it
			if srv.Address == raft.ServerAddress(bindAddr) && srv.ID == raft.ServerID(nodeID) {
				return nil
			}

			future := s.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return err
			}
		}
	}

	f := s.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(bindAddr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}

	return nil
}

// Leader is a convienence function exposing if we are leader or not
func (s *Store) IsLeader() bool {
	return s.raft.State() == raft.Leader
}
