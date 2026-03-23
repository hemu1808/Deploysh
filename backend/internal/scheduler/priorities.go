package scheduler

import "github.com/hemu1808/auradeploy/backend/internal/models"

type Priority func(node *models.Node, app *models.Application, placements map[string][]string, pvcs map[string]models.PersistentVolumeClaim, pvs map[string]models.PersistentVolume) int

// LeastAllocated scores nodes higher if they have fewer placements
func LeastAllocated(node *models.Node, app *models.Application, placementsByNode map[string][]string, _ map[string]models.PersistentVolumeClaim, _ map[string]models.PersistentVolume) int {
	// placementsByNode maps NodeID to a list of app IDs placed there.
	count := len(placementsByNode[node.ID])
	// Lowest count gets highest score (e.g., 100 - count)
	score := 100 - count
	if score < 0 {
		return 0
	}
	return score
}
