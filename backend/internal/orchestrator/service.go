package orchestrator

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hemu1808/auradeploy/backend/internal/cri"
	"github.com/hemu1808/auradeploy/backend/internal/models"
	"github.com/hemu1808/auradeploy/backend/internal/store"
	"log/slog"
)

type Orchestrator interface {
	GetApplications() ([]models.Application, error)
	DeployApplication(image string, replicas int, envVars []models.EnvVar, ports []models.PortMapping) (models.Application, error)
	ScaleApplication(appID string, targetReplicas int) (models.Application, error)
	RemoveApplication(appID string) error
	JoinCluster(nodeID, bindAddr string) error
	Subscribe(subscriber chan []models.Application)
	Unsubscribe(subscriber chan []models.Application)
}

func (o *RaftOrchestrator) JoinCluster(nodeID, bindAddr string) error {
	return o.store.Join(nodeID, bindAddr)
}

type RaftOrchestrator struct {
	mu          sync.RWMutex
	store       *store.Store
	criClient   *cri.Client
	subscribers map[chan []models.Application]struct{}
	logger      *slog.Logger
	running     map[string]models.Application // keeps track of what this node is actually running
}

func NewRaftOrchestrator(store *store.Store, criClient *cri.Client, logger *slog.Logger) *RaftOrchestrator {
	orch := &RaftOrchestrator{
		store:       store,
		criClient:   criClient,
		subscribers: make(map[chan []models.Application]struct{}),
		running:     make(map[string]models.Application),
		logger:      logger,
	}
	
	go orch.watchStateLoop()
	return orch
}

func (o *RaftOrchestrator) GetApplications() ([]models.Application, error) {
	appMap := o.store.GetApplications()
	var apps []models.Application
	for _, app := range appMap {
		apps = append(apps, app)
	}
	return apps, nil
}

func (o *RaftOrchestrator) DeployApplication(image string, replicas int, envVars []models.EnvVar, ports []models.PortMapping) (models.Application, error) {
	if !o.store.IsLeader() {
		return models.Application{}, errors.New("cannot deploy: not the raft leader")
	}

	appID := "app-" + time.Now().Format("20060102150405")
	name := "service-" + appID[len(appID)-4:]

	app := models.Application{
		ID:          appID,
		Name:        name,
		DockerImage: image,
		Status:      models.StatusDeploying,
		Replicas: models.ReplicasJSON{
			Current: 0,
			Target:  replicas,
		},
		EnvVars:     models.EnvVarsJSON(envVars),
		Ports:       models.PortsJSON(ports),
		Logs:        models.LogsJSON{"[INFO] Deployment initiated. Proposing to Raft cluster..."},
		CPUUsage:    models.MetricsJSON{},
		MemoryUsage: models.MetricsJSON{},
	}

	err := o.store.SetApplication(app)
	if err != nil {
		o.logger.Error("Failed to propose application to Raft", "error", err)
		return models.Application{}, err
	}

	o.logger.Info("Deployed new application via Raft", "id", appID, "image", image)
	o.broadcast()

	return app, nil
}

func (o *RaftOrchestrator) ScaleApplication(appID string, targetReplicas int) (models.Application, error) {
	if !o.store.IsLeader() {
		return models.Application{}, errors.New("cannot scale: not the raft leader")
	}

	apps := o.store.GetApplications()
	app, exists := apps[appID]
	if !exists {
		return models.Application{}, errors.New("application not found")
	}

	// Update replicas specifically
	replicas := models.Replicas(app.Replicas)
	replicas.Target = targetReplicas
	
	app.Replicas = models.ReplicasJSON(replicas)
	app.Status = models.StatusDeploying
	
	err := o.store.SetApplication(app)
	if err != nil {
		return models.Application{}, err
	}
	
	o.logger.Info("Scaled application via Raft", "id", appID, "target", targetReplicas)
	o.broadcast()

	return app, nil
}

func (o *RaftOrchestrator) RemoveApplication(appID string) error {
	if !o.store.IsLeader() {
		return errors.New("cannot remove: not the raft leader")
	}

	err := o.store.DeleteApplication(appID)
	if err != nil {
		return err
	}

	o.logger.Info("Removed application via Raft", "id", appID)
	o.broadcast()

	return nil
}

func (o *RaftOrchestrator) Subscribe(ch chan []models.Application) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.subscribers[ch] = struct{}{}
	o.logger.Debug("New websocket client subscribed")
}

func (o *RaftOrchestrator) Unsubscribe(ch chan []models.Application) {
	o.mu.Lock()
	defer o.mu.Unlock()
	delete(o.subscribers, ch)
	close(ch)
	o.logger.Debug("Websocket client unsubscribed")
}

func (o *RaftOrchestrator) broadcast() {
	apps, err := o.GetApplications()
	if err != nil {
		o.logger.Error("Failed to fetch apps for broadcast", "error", err)
		return
	}

	o.mu.RLock()
	defer o.mu.RUnlock()
	
	for ch := range o.subscribers {
		select {
		case ch <- apps:
		default:
		}
	}
}

// watchStateLoop reconciles the local containerd state with the global Raft state
func (o *RaftOrchestrator) watchStateLoop() {
	ticker := time.NewTicker(2 * time.Second)
	
	for range ticker.C {
		apps := o.store.GetApplications()
		
		// 1. Check for new apps to run locally
		// In a real system the scheduler assigns nodes. Here we blindly run everything.
		for id, app := range apps {
			if _, running := o.running[id]; !running {
				o.logger.Info("New app detected in state, spinning up via CRI", "id", id)
				
				// Optional: Set status to Deploying via raft if we are leader
				
				if o.criClient != nil {
					err := o.criClient.RunContainer(app)
					if err != nil {
						o.logger.Error("Failed to run container", "id", id, "error", err)
						// Update Raft state
						if o.store.IsLeader() {
							app.Status = models.StatusUnhealthy
							app.Logs = append(app.Logs, fmt.Sprintf("[ERROR] CRI failed to run: %v", err))
							o.store.SetApplication(app)
						}
					} else {
						o.running[id] = app
						// Update Raft state
						if o.store.IsLeader() {
							app.Status = models.StatusHealthy
							replicas := models.Replicas(app.Replicas)
							replicas.Current = replicas.Target
							app.Replicas = models.ReplicasJSON(replicas)
							app.Logs = append(app.Logs, "[INFO] Container started via CRI")
							o.store.SetApplication(app)
						}
					}
				} else {
					// Fallback for demo without containerd
					o.running[id] = app
					if o.store.IsLeader() {
						app.Status = models.StatusHealthy
						replicas := models.Replicas(app.Replicas)
						replicas.Current = replicas.Target
						app.Replicas = models.ReplicasJSON(replicas)
						app.Logs = append(app.Logs, "[INFO] Simulated container started.")
						o.store.SetApplication(app)
					}
				}
			} else {
				// Container is running, collect real metrics
				if o.criClient != nil && o.store.IsLeader() {
					cpu, mem, err := o.criClient.GetMetrics(id)
					if err == nil {
						now := time.Now().UnixMilli()
						app.CPUUsage = append(app.CPUUsage, models.Metric{Time: now, Value: float64(cpu)})
						app.MemoryUsage = append(app.MemoryUsage, models.Metric{Time: now, Value: float64(mem)})
						
						if len(app.CPUUsage) > 30 { app.CPUUsage = app.CPUUsage[1:] }
						if len(app.MemoryUsage) > 30 { app.MemoryUsage = app.MemoryUsage[1:] }
						
						o.store.SetApplication(app)
					}
				}
			}
		}

		// 2. Check for apps deleted from state that are still running locally
		for id := range o.running {
			if _, exists := apps[id]; !exists {
				o.logger.Info("App deleted from state, stopping via CRI", "id", id)
				if o.criClient != nil {
					err := o.criClient.StopContainer(id)
					if err != nil {
						o.logger.Error("Failed to stop container", "id", id, "error", err)
					}
				}
				delete(o.running, id)
			}
		}

		o.broadcast()
	}
}
