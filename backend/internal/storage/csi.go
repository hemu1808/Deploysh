package storage

import (
	"log/slog"
	"time"

	"github.com/hemu1808/auradeploy/backend/internal/models"
	"github.com/hemu1808/auradeploy/backend/internal/store"
)

type Provisioner struct {
	store   *store.Store
	logger  *slog.Logger
	dataDir string
}

func NewProvisioner(s *store.Store, logger *slog.Logger, dataDir string) *Provisioner {
	return &Provisioner{
		store:   s,
		logger:  logger,
		dataDir: dataDir,
	}
}

func (p *Provisioner) Start() {
	ticker := time.NewTicker(3 * time.Second)
	go func() {
		for range ticker.C {
			if p.store.IsLeader() {
				p.reconcilePVCs()
			}
		}
	}()
}

func (p *Provisioner) reconcilePVCs() {
	pvcs := p.store.GetPVCs()
	nodes := p.store.GetNodes()

	if len(nodes) == 0 {
		return
	}

	var nodeIDs []string
	for id := range nodes {
		nodeIDs = append(nodeIDs, id)
	}

	for _, pvc := range pvcs {
		if pvc.Phase == models.VolumePending {
			// Basic static provisioning: assign to the first available node
			nodeID := nodeIDs[0]

			pvID := "pv-" + pvc.ID
			pv := models.PersistentVolume{
				ID:        pvID,
				ClaimID:   pvc.ID,
				NodeID:    nodeID,
				HostPath:  p.dataDir + "/" + pvID,
				SizeMB:    pvc.SizeMB,
				Phase:     models.VolumeBound,
				CreatedAt: time.Now().UnixMilli(),
			}

			if err := p.store.SetPV(pv); err != nil {
				p.logger.Error("Failed to provision PV", "error", err)
				continue
			}

			pvc.Phase = models.VolumeBound
			pvc.VolumeID = pvID
			if err := p.store.SetPVC(pvc); err != nil {
				p.logger.Error("Failed to update PVC", "error", err)
			} else {
				p.logger.Info("Provisioned local PV for PVC", "pvc", pvc.ID, "pv", pv.ID, "node", nodeID)
			}
		}
	}
}
