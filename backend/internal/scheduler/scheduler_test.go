package scheduler

import (
	"testing"

	"github.com/hemu1808/auradeploy/backend/internal/models"
)

func TestHasSufficientResources(t *testing.T) {
	app := &models.Application{}

	t.Run("Ready Node with resources", func(t *testing.T) {
		node := &models.Node{
			Status: models.NodeStatusReady,
			Capacity: models.ResourceCapacity{
				MemoryMB: 1024,
				CPUCores: 2,
			},
		}
		if !HasSufficientResources(node, app, nil, nil) {
			t.Errorf("Expected node to have sufficient resources")
		}
	})

	t.Run("NotReady Node", func(t *testing.T) {
		node := &models.Node{
			Status: models.NodeStatusNotReady,
			Capacity: models.ResourceCapacity{
				MemoryMB: 1024,
				CPUCores: 2,
			},
		}
		if HasSufficientResources(node, app, nil, nil) {
			t.Errorf("Expected NotReady node to fail check")
		}
	})

	t.Run("Zero Capacity Node", func(t *testing.T) {
		node := &models.Node{
			Status:   models.NodeStatusReady,
			Capacity: models.ResourceCapacity{MemoryMB: 0, CPUCores: 0},
		}
		if HasSufficientResources(node, app, nil, nil) {
			t.Errorf("Expected zero capacity node to fail check")
		}
	})
}

func TestLeastAllocated(t *testing.T) {
	app := &models.Application{ID: "app1"}

	nodeA := &models.Node{ID: "nodeA"}
	nodeB := &models.Node{ID: "nodeB"}

	placements := map[string][]string{
		"nodeA": {"app2", "app3"}, // 2 placements
		"nodeB": {"app4"},         // 1 placement
	}

	scoreA := LeastAllocated(nodeA, app, placements, nil, nil)
	scoreB := LeastAllocated(nodeB, app, placements, nil, nil)

	if scoreB <= scoreA {
		t.Errorf("Expected nodeB (1 placement) to score higher than nodeA (2 placements). Got A: %d, B: %d", scoreA, scoreB)
	}
}
