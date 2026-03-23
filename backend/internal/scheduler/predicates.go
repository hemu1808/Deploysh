package scheduler

import "github.com/hemu1808/auradeploy/backend/internal/models"

type Predicate func(node *models.Node, app *models.Application, pvcs map[string]models.PersistentVolumeClaim, pvs map[string]models.PersistentVolume) bool

// HasSufficientResources checks if the node has enough CPU and Memory
func HasSufficientResources(node *models.Node, app *models.Application, _ map[string]models.PersistentVolumeClaim, _ map[string]models.PersistentVolume) bool {
	// For simplicity, we just check if node is Ready and has non-zero capacity.
	// In a real bin-packing scenario, we would subtract resources used by existing placements.
	if node.Status != models.NodeStatusReady {
		return false
	}
	return node.Capacity.MemoryMB > 0 && node.Capacity.CPUCores > 0
}

// VolumeNodeAffinity checks if the Application requires a volume, and if it's already bound to a different node
func VolumeNodeAffinity(node *models.Node, app *models.Application, pvcs map[string]models.PersistentVolumeClaim, pvs map[string]models.PersistentVolume) bool {
	for _, mount := range app.VolumeMounts {
		pvc, exists := pvcs[mount.ClaimID]
		if !exists {
			return false // Requires a PVC that doesn't exist yet
		}
		
		if pvc.Phase != models.VolumeBound {
			return false // Can't schedule until volume is bound and provisioned
		}
		
		pv, exists := pvs[pvc.VolumeID]
		if !exists {
			return false
		}
		
		if pv.NodeID != node.ID {
			return false // App relies on a local volume provisioned on a different node
		}
	}
	return true
}
