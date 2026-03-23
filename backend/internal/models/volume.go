package models

import "time"

// VolumePhase defines the lifecycle of a PVC or PV
type VolumePhase string

const (
	VolumePending VolumePhase = "Pending"
	VolumeBound   VolumePhase = "Bound"
	VolumeFailed  VolumePhase = "Failed"
)

// PersistentVolumeClaim represents a user request for storage
type PersistentVolumeClaim struct {
	ID        string      `json:"id"`
	AppID     string      `json:"appId"` // The app that requested it
	SizeMB    int         `json:"sizeMb"`
	Phase     VolumePhase `json:"phase"`
	VolumeID  string      `json:"volumeId,omitempty"` // Bound PV ID
	CreatedAt int64       `json:"createdAt"`
}

// PersistentVolume represents the actual provisioned storage on a Node
type PersistentVolume struct {
	ID        string      `json:"id"`
	ClaimID   string      `json:"claimId"`
	NodeID    string      `json:"nodeId"`   // Node Affinity
	HostPath  string      `json:"hostPath"` // Where it actually lives on the node
	SizeMB    int         `json:"sizeMb"`
	Phase     VolumePhase `json:"phase"`
	CreatedAt int64       `json:"createdAt"`
}

func NewPVC(id, appID string, sizeMb int) PersistentVolumeClaim {
	return PersistentVolumeClaim{
		ID:        id,
		AppID:     appID,
		SizeMB:    sizeMb,
		Phase:     VolumePending,
		CreatedAt: time.Now().UnixMilli(),
	}
}
