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

// GetNodes returns a copy of all nodes in the FSM state
func (s *Store) GetNodes() map[string]models.Node {
	s.fsm.mu.RLock()
	defer s.fsm.mu.RUnlock()

	clone := make(map[string]models.Node)
	for k, v := range s.fsm.nodes {
		clone[k] = v
	}

	return clone
}

// GetPVCs returns a copy of all PVCs in the FSM state
func (s *Store) GetPVCs() map[string]models.PersistentVolumeClaim {
	s.fsm.mu.RLock()
	defer s.fsm.mu.RUnlock()

	clone := make(map[string]models.PersistentVolumeClaim)
	for k, v := range s.fsm.pvcs {
		clone[k] = v
	}

	return clone
}

// GetPVs returns a copy of all PVs in the FSM state
func (s *Store) GetPVs() map[string]models.PersistentVolume {
	s.fsm.mu.RLock()
	defer s.fsm.mu.RUnlock()

	clone := make(map[string]models.PersistentVolume)
	for k, v := range s.fsm.pvs {
		clone[k] = v
	}

	return clone
}

// GetRoles returns a copy of all Roles in the FSM state
func (s *Store) GetRoles() map[string]models.Role {
	s.fsm.mu.RLock()
	defer s.fsm.mu.RUnlock()

	clone := make(map[string]models.Role)
	for k, v := range s.fsm.roles {
		clone[k] = v
	}

	return clone
}

// GetRoleBindings returns a copy of all RoleBindings in the FSM state
func (s *Store) GetRoleBindings() map[string]models.RoleBinding {
	s.fsm.mu.RLock()
	defer s.fsm.mu.RUnlock()

	clone := make(map[string]models.RoleBinding)
	for k, v := range s.fsm.roleBindings {
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

// SetNode proposes a state change to Raft: creating or replacing a Node
func (s *Store) SetNode(node models.Node) error {
	if s.raft.State() != raft.Leader {
		return fmt.Errorf("not the leader")
	}

	c := command{
		Type: SetNodeCommand,
		Node: node,
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

// DeleteNode proposes a state change to Raft: removing a Node
func (s *Store) DeleteNode(nodeID string) error {
	if s.raft.State() != raft.Leader {
		return fmt.Errorf("not the leader")
	}

	c := command{
		Type: DeleteNodeCommand,
		ID:   nodeID,
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

// Volume Commands

func (s *Store) SetPVC(pvc models.PersistentVolumeClaim) error {
	if s.raft.State() != raft.Leader {
		return fmt.Errorf("not the leader")
	}
	c := command{Type: SetPVCCommand, PVC: pvc}
	b, _ := json.Marshal(c)
	if err := s.raft.Apply(b, 5*time.Second).Error(); err != nil {
		return err
	}
	return nil
}

func (s *Store) DeletePVC(id string) error {
	if s.raft.State() != raft.Leader {
		return fmt.Errorf("not the leader")
	}
	c := command{Type: DeletePVCCommand, ID: id}
	b, _ := json.Marshal(c)
	if err := s.raft.Apply(b, 5*time.Second).Error(); err != nil {
		return err
	}
	return nil
}

func (s *Store) SetPV(pv models.PersistentVolume) error {
	if s.raft.State() != raft.Leader {
		return fmt.Errorf("not the leader")
	}
	c := command{Type: SetPVCommand, PV: pv}
	b, _ := json.Marshal(c)
	if err := s.raft.Apply(b, 5*time.Second).Error(); err != nil {
		return err
	}
	return nil
}

func (s *Store) DeletePV(id string) error {
	if s.raft.State() != raft.Leader {
		return fmt.Errorf("not the leader")
	}
	c := command{Type: DeletePVCommand, ID: id}
	b, _ := json.Marshal(c)
	if err := s.raft.Apply(b, 5*time.Second).Error(); err != nil {
		return err
	}
	return nil
}

// RBAC Commands

func (s *Store) SetRole(role models.Role) error {
	if s.raft.State() != raft.Leader {
		return fmt.Errorf("not the leader")
	}
	c := command{Type: SetRoleCommand, Role: role}
	b, _ := json.Marshal(c)
	if err := s.raft.Apply(b, 5*time.Second).Error(); err != nil {
		return err
	}
	return nil
}

func (s *Store) DeleteRole(id string) error {
	if s.raft.State() != raft.Leader {
		return fmt.Errorf("not the leader")
	}
	c := command{Type: DeleteRoleCommand, ID: id}
	b, _ := json.Marshal(c)
	if err := s.raft.Apply(b, 5*time.Second).Error(); err != nil {
		return err
	}
	return nil
}

func (s *Store) SetRoleBinding(rb models.RoleBinding) error {
	if s.raft.State() != raft.Leader {
		return fmt.Errorf("not the leader")
	}
	c := command{Type: SetRoleBindingCommand, RoleBinding: rb}
	b, _ := json.Marshal(c)
	if err := s.raft.Apply(b, 5*time.Second).Error(); err != nil {
		return err
	}
	return nil
}

func (s *Store) DeleteRoleBinding(id string) error {
	if s.raft.State() != raft.Leader {
		return fmt.Errorf("not the leader")
	}
	c := command{Type: DeleteRoleBindingCommand, ID: id}
	b, _ := json.Marshal(c)
	if err := s.raft.Apply(b, 5*time.Second).Error(); err != nil {
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
