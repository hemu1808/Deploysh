package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrPrivilegedContainer = errors.New("PodSecurity Standard Violation: Privileged containers (RUN_AS_ROOT) are not allowed")
	ErrPrivilegedPort      = errors.New("PodSecurity Standard Violation: Binding to host ports < 1024 is not allowed")
)

// ApplicationStatus defines the current state of an application
type ApplicationStatus string

const (
	StatusHealthy   ApplicationStatus = "Healthy"
	StatusUnhealthy ApplicationStatus = "Unhealthy"
	StatusDeploying ApplicationStatus = "Deploying"
	StatusStopped   ApplicationStatus = "Stopped"
)

// Metric represents a time-series data point (CPU/Memory)
type Metric struct {
	Time  int64   `json:"time"`
	Value float64 `json:"value"`
}

// EnvVar represents a single environment variable key-value pair
type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PortMapping represents a host port to container port binding
type PortMapping struct {
	HostPort      int `json:"hostPort"`
	ContainerPort int `json:"containerPort"`
}

// VolumeMount represents a PVC mounted into a container
type VolumeMount struct {
	ClaimID   string `json:"claimId"`
	MountPath string `json:"mountPath"`
}

// Replicas defines the current vs desired instances
type Replicas struct {
	Current int `json:"current"`
	Target  int `json:"target"`
}

// Application is the core domain model representing a deployed service
type Application struct {
	ID          string            `gorm:"primaryKey" json:"id"`
	Name        string            `json:"name"`
	DockerImage string            `json:"dockerImage"`
	Status      ApplicationStatus `json:"status"`
	
	// Complex JSON columns
	Replicas    ReplicasJSON      `gorm:"type:text" json:"replicas"`
	CPUUsage    MetricsJSON       `gorm:"type:text" json:"cpuUsage"`
	MemoryUsage MetricsJSON       `gorm:"type:text" json:"memoryUsage"`
	Logs        LogsJSON          `gorm:"type:text" json:"logs"`
	EnvVars     EnvVarsJSON       `gorm:"type:text" json:"envVars,omitempty"`
	Ports       PortsJSON         `gorm:"type:text" json:"ports,omitempty"`
	Placements  map[string]string `gorm:"type:text" json:"placements,omitempty"` // replicaID -> NodeID
	VolumeMounts VolumeMountsJSON `gorm:"type:text" json:"volumeMounts,omitempty"`
	
	CreatedAt   int64           `json:"createdAt"`
	UpdatedAt   int64           `json:"-"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"-"`
}

// Implement hooks to set times automatically
func (a *Application) BeforeCreate(tx *gorm.DB) (err error) {
	if a.CreatedAt == 0 {
		a.CreatedAt = time.Now().UnixMilli()
	}
	a.UpdatedAt = time.Now().UnixMilli()
	return
}

func (a *Application) BeforeUpdate(tx *gorm.DB) (err error) {
	a.UpdatedAt = time.Now().UnixMilli()
	return
}

// Custom Types mapping to Text JSON columns for GORM

type ReplicasJSON Replicas
func (r ReplicasJSON) Value() (driver.Value, error) { return json.Marshal(r) }
func (r *ReplicasJSON) Scan(value interface{}) error { return scanJSON(value, r) }

type MetricsJSON []Metric
func (m MetricsJSON) Value() (driver.Value, error) { return json.Marshal(m) }
func (m *MetricsJSON) Scan(value interface{}) error { return scanJSON(value, m) }

type LogsJSON []string
func (l LogsJSON) Value() (driver.Value, error) { return json.Marshal(l) }
func (l *LogsJSON) Scan(value interface{}) error { return scanJSON(value, l) }

type EnvVarsJSON []EnvVar
func (e EnvVarsJSON) Value() (driver.Value, error) { return json.Marshal(e) }
func (e *EnvVarsJSON) Scan(value interface{}) error { return scanJSON(value, e) }

type PortsJSON []PortMapping
func (p PortsJSON) Value() (driver.Value, error) { return json.Marshal(p) }
func (p *PortsJSON) Scan(value interface{}) error { return scanJSON(value, p) }

type VolumeMountsJSON []VolumeMount
func (v VolumeMountsJSON) Value() (driver.Value, error) { return json.Marshal(v) }
func (v *VolumeMountsJSON) Scan(value interface{}) error { return scanJSON(value, v) }

// Generic JSON scanner
func scanJSON(value interface{}, dest interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		// SQLite might return a string
		str, ok := value.(string)
		if !ok {
			return errors.New("failed to unmarshal JSON value: not byte array or string")
		}
		bytes = []byte(str)
	}
	
	if len(bytes) == 0 {
		// Keep dest initialized as empty
		return nil
	}
	
	return json.Unmarshal(bytes, dest)
}
