package models

import "time"

// NodeStatus defines the current state of a cluster node
type NodeStatus string

const (
	NodeStatusReady    NodeStatus = "Ready"
	NodeStatusNotReady NodeStatus = "NotReady"
)

// Node represents a machine/worker in the cluster
type Node struct {
	ID        string            `json:"id"`
	Address   string            `json:"address"`
	Status    NodeStatus        `json:"status"`
	Labels    map[string]string `json:"labels"`
	Capacity  ResourceCapacity  `json:"capacity"`
	JoinedAt  int64             `json:"joinedAt"`
	UpdatedAt int64             `json:"updatedAt"`
}

// ResourceCapacity defines the resources a node has available
type ResourceCapacity struct {
	MemoryMB int `json:"memoryMb"`
	CPUCores int `json:"cpuCores"`
}

// DefaultNode creates a basic node definition
func DefaultNode(id, address string) Node {
	return Node{
		ID:      id,
		Address: address,
		Status:  NodeStatusReady,
		Labels:  map[string]string{"kubernetes.io/hostname": id}, // Mimicking K8s
		Capacity: ResourceCapacity{
			MemoryMB: 4096, // placeholder
			CPUCores: 4,    // placeholder
		},
		JoinedAt:  time.Now().UnixMilli(),
		UpdatedAt: time.Now().UnixMilli(),
	}
}
