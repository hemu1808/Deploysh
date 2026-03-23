package scheduler

import (
	"log/slog"
	"sort"
	"time"

	"github.com/hemu1808/auradeploy/backend/internal/models"
	"github.com/hemu1808/auradeploy/backend/internal/store"
)

type Scheduler struct {
	store  *store.Store
	logger *slog.Logger

	predicates []Predicate
	priorities []Priority
}

func NewScheduler(s *store.Store, logger *slog.Logger) *Scheduler {
	return &Scheduler{
		store:      s,
		logger:     logger,
		predicates: []Predicate{HasSufficientResources, VolumeNodeAffinity},
		priorities: []Priority{LeastAllocated},
	}
}

func (s *Scheduler) Start() {
	ticker := time.NewTicker(3 * time.Second)
	go func() {
		for range ticker.C {
			if s.store.IsLeader() {
				s.scheduleLoop()
			}
		}
	}()
}

func (s *Scheduler) scheduleLoop() {
	apps := s.store.GetApplications()
	nodes := s.store.GetNodes()
	pvcs := s.store.GetPVCs()
	pvs := s.store.GetPVs()

	// Build placementsByNode
	placementsByNode := make(map[string][]string)
	for _, app := range apps {
		for _, nID := range app.Placements {
			placementsByNode[nID] = append(placementsByNode[nID], app.ID)
		}
	}

	// Attempt to schedule unplaced replicas
	for _, app := range apps {
		// If app.Placements is nil, initialize it
		if app.Placements == nil {
			app.Placements = make(map[string]string)
		}

		// Count how many we have vs Target
		currentPlacements := len(app.Placements)
		target := app.Replicas.Target

		if currentPlacements < target {
			s.logger.Info("App needs scheduling", "app", app.ID, "current", currentPlacements, "target", target)

			// Loop for each missing replica
			for i := currentPlacements; i < target; i++ {
				bestNode := s.findBestNode(&app, nodes, placementsByNode, pvcs, pvs)
				if bestNode != "" {
					replicaID := app.ID + "-rep-" + time.Now().Format("150405.000000") // fake unique ID
					s.logger.Info("Scheduled replica to node", "app", app.ID, "replica", replicaID, "node", bestNode)
					app.Placements[replicaID] = bestNode
					placementsByNode[bestNode] = append(placementsByNode[bestNode], app.ID)

					// Propose state update immediately
					err := s.store.SetApplication(app)
					if err != nil {
						s.logger.Error("Failed to save placement", "error", err)
					}
				} else {
					s.logger.Warn("Failed to find suitable node for replica", "app", app.ID)
				}
			}
		} else if currentPlacements > target {
			// Handle scale down: remove excess placements
			s.logger.Info("App needs scaling down", "app", app.ID, "current", currentPlacements, "target", target)
			diff := currentPlacements - target
			removed := 0

			for repID, nodeID := range app.Placements {
				if removed >= diff {
					break
				}
				s.logger.Info("Removing replica", "app", app.ID, "replica", repID, "node", nodeID)
				delete(app.Placements, repID)
				removed++
			}

			err := s.store.SetApplication(app)
			if err != nil {
				s.logger.Error("Failed to save downscaled placement", "error", err)
			}
		}
	}
}

func (s *Scheduler) findBestNode(app *models.Application, nodes map[string]models.Node, placementsByNode map[string][]string, pvcs map[string]models.PersistentVolumeClaim, pvs map[string]models.PersistentVolume) string {
	var validNodes []*models.Node

	// 1. Run Predicates
	for _, node := range nodes {
		n := node // copy
		valid := true
		for _, p := range s.predicates {
			if !p(&n, app, pvcs, pvs) {
				valid = false
				break
			}
		}
		if valid {
			validNodes = append(validNodes, &n)
		}
	}

	if len(validNodes) == 0 {
		return ""
	}

	// 2. Run Priorities
	type nodeScore struct {
		nodeID string
		score  int
	}

	var scores []nodeScore
	for _, n := range validNodes {
		totalScore := 0
		for _, p := range s.priorities {
			totalScore += p(n, app, placementsByNode, pvcs, pvs)
		}
		scores = append(scores, nodeScore{nodeID: n.ID, score: totalScore})
	}

	// 3. Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	return scores[0].nodeID
}
