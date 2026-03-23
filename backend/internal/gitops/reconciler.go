package gitops

import (
	"log/slog"
	"reflect"

	"github.com/hemu1808/auradeploy/backend/internal/models"
	"github.com/hemu1808/auradeploy/backend/internal/orchestrator"
)

type Reconciler struct {
	orch   orchestrator.Orchestrator
	logger *slog.Logger
}

func NewReconciler(orch orchestrator.Orchestrator, logger *slog.Logger) *Reconciler {
	return &Reconciler{orch: orch, logger: logger}
}

// Reconcile takes a list of desired applications and applies drift updates to the cluster
func (r *Reconciler) Reconcile(desired []models.Application) error {
	actual, err := r.orch.GetApplications()
	if err != nil {
		return err
	}

	actualMap := make(map[string]models.Application)
	for _, a := range actual {
		actualMap[a.ID] = a
	}

	for _, reqApp := range desired {
		curr, exists := actualMap[reqApp.ID]
		if !exists {
			r.logger.Info("GitOps: Deploying new application", "app", reqApp.ID)
			_, err := r.orch.DeployApplication(reqApp.DockerImage, reqApp.Replicas.Target, reqApp.EnvVars, reqApp.Ports)
			if err != nil {
				r.logger.Error("GitOps Deploy failed", "error", err)
			}
			continue
		}

		// Check for drift (Image or Replicas or EnvVars diff)
		drift := false
		if curr.DockerImage != reqApp.DockerImage {
			drift = true
			r.logger.Info("GitOps Drift Detected: Image", "app", reqApp.ID, "old", curr.DockerImage, "new", reqApp.DockerImage)
		}
		if curr.Replicas.Target != reqApp.Replicas.Target {
			drift = true
			r.logger.Info("GitOps Drift Detected: Replicas", "app", reqApp.ID, "old", curr.Replicas.Target, "new", reqApp.Replicas.Target)
		}
		if !reflect.DeepEqual(curr.EnvVars, reqApp.EnvVars) {
			drift = true
			r.logger.Info("GitOps Drift Detected: EnvVars", "app", reqApp.ID)
		}

		if drift {
			r.logger.Info("GitOps: Reconciling Drift", "app", reqApp.ID)
			// Our DeployApplication logic acts as a Create-or-Replace
			_, err := r.orch.DeployApplication(reqApp.DockerImage, reqApp.Replicas.Target, reqApp.EnvVars, reqApp.Ports)
			if err != nil {
				r.logger.Error("GitOps Reconciliation failed", "error", err)
			}
		}
		
		// Mark as seen
		delete(actualMap, reqApp.ID)
	}

	// Leftovers in actualMap are orphaned. 
	// In strict mode we'd prune them automatically, but here we just warn.
	for id := range actualMap {
		r.logger.Warn("GitOps: Application exists in cluster but not in Git (Drift)", "app", id)
	}

	return nil
}
