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
	mu           sync.RWMutex
	apps         map[string]models.Application
	nodes        map[string]models.Node
	pvcs         map[string]models.PersistentVolumeClaim
	pvs          map[string]models.PersistentVolume
	roles        map[string]models.Role
	roleBindings map[string]models.RoleBinding
}

func newFSM() *fsm {
	return &fsm{
		apps:         make(map[string]models.Application),
		nodes:        make(map[string]models.Node),
		pvcs:         make(map[string]models.PersistentVolumeClaim),
		pvs:          make(map[string]models.PersistentVolume),
		roles:        make(map[string]models.Role),
		roleBindings: make(map[string]models.RoleBinding),
	}
}

// Commands
type commandType string

const (
	SetAppCommand         commandType = "SET_APP"
	DeleteAppCommand      commandType = "DELETE_APP"
	SetNodeCommand        commandType = "SET_NODE"
	DeleteNodeCommand     commandType = "DELETE_NODE"
	SetPVCCommand         commandType = "SET_PVC"
	DeletePVCCommand      commandType = "DELETE_PVC"
	SetPVCommand          commandType = "SET_PV"
	DeletePVCommand       commandType = "DELETE_PV"
	SetRoleCommand        commandType = "SET_ROLE"
	DeleteRoleCommand     commandType = "DELETE_ROLE"
	SetRoleBindingCommand commandType = "SET_ROLE_BINDING"
	DeleteRoleBindingCommand commandType = "DELETE_ROLE_BINDING"
)

type command struct {
	Type        commandType                  `json:"type"`
	App         models.Application           `json:"app,omitempty"`
	Node        models.Node                  `json:"node,omitempty"`
	PVC         models.PersistentVolumeClaim `json:"pvc,omitempty"`
	PV          models.PersistentVolume      `json:"pv,omitempty"`
	Role        models.Role                  `json:"role,omitempty"`
	RoleBinding models.RoleBinding           `json:"roleBinding,omitempty"`
	ID          string                       `json:"id,omitempty"`
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
	case SetNodeCommand:
		f.nodes[c.Node.ID] = c.Node
	case DeleteNodeCommand:
		delete(f.nodes, c.ID)
	case SetPVCCommand:
		f.pvcs[c.PVC.ID] = c.PVC
	case DeletePVCCommand:
		delete(f.pvcs, c.ID)
	case SetPVCommand:
		f.pvs[c.PV.ID] = c.PV
	case DeletePVCommand:
		delete(f.pvs, c.ID)
	case SetRoleCommand:
		f.roles[c.Role.ID] = c.Role
	case DeleteRoleCommand:
		delete(f.roles, c.ID)
	case SetRoleBindingCommand:
		f.roleBindings[c.RoleBinding.ID] = c.RoleBinding
	case DeleteRoleBindingCommand:
		delete(f.roleBindings, c.ID)
	}

	return nil
}

// Snapshot returns a snapshot of the key-value store.
func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Clone the map
	cloneApps := make(map[string]models.Application)
	for k, v := range f.apps {
		cloneApps[k] = v
	}
	
	cloneNodes := make(map[string]models.Node)
	for k, v := range f.nodes {
		cloneNodes[k] = v
	}

	clonePVCs := make(map[string]models.PersistentVolumeClaim)
	for k, v := range f.pvcs {
		clonePVCs[k] = v
	}
	clonePVs := make(map[string]models.PersistentVolume)
	for k, v := range f.pvs {
		clonePVs[k] = v
	}
	cloneRoles := make(map[string]models.Role)
	for k, v := range f.roles {
		cloneRoles[k] = v
	}
	cloneRoleBindings := make(map[string]models.RoleBinding)
	for k, v := range f.roleBindings {
		cloneRoleBindings[k] = v
	}

	return &fsmSnapshot{
		apps:         cloneApps,
		nodes:        cloneNodes,
		pvcs:         clonePVCs,
		pvs:          clonePVs,
		roles:        cloneRoles,
		roleBindings: cloneRoleBindings,
	}, nil
}

// Restore stores the key-value store to a previous state.
func (f *fsm) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	var state struct {
		Apps         map[string]models.Application           `json:"apps"`
		Nodes        map[string]models.Node                  `json:"nodes"`
		PVCs         map[string]models.PersistentVolumeClaim `json:"pvcs"`
		PVs          map[string]models.PersistentVolume      `json:"pvs"`
		Roles        map[string]models.Role                  `json:"roles"`
		RoleBindings map[string]models.RoleBinding           `json:"roleBindings"`
	}
	if err := json.NewDecoder(rc).Decode(&state); err != nil {
		return err
	}

	f.mu.Lock()
	f.apps = state.Apps
	f.nodes = state.Nodes
	f.pvcs = state.PVCs
	f.pvs = state.PVs
	f.roles = state.Roles
	f.roleBindings = state.RoleBindings
	f.mu.Unlock()

	return nil
}

type fsmSnapshot struct {
	apps         map[string]models.Application
	nodes        map[string]models.Node
	pvcs         map[string]models.PersistentVolumeClaim
	pvs          map[string]models.PersistentVolume
	roles        map[string]models.Role
	roleBindings map[string]models.RoleBinding
}

func (s *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		state := struct {
			Apps         map[string]models.Application           `json:"apps"`
			Nodes        map[string]models.Node                  `json:"nodes"`
			PVCs         map[string]models.PersistentVolumeClaim `json:"pvcs"`
			PVs          map[string]models.PersistentVolume      `json:"pvs"`
			Roles        map[string]models.Role                  `json:"roles"`
			RoleBindings map[string]models.RoleBinding           `json:"roleBindings"`
		}{
			Apps:         s.apps,
			Nodes:        s.nodes,
			PVCs:         s.pvcs,
			PVs:          s.pvs,
			Roles:        s.roles,
			RoleBindings: s.roleBindings,
		}

		
		b, err := json.Marshal(state)
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
