package api

import (
	"encoding/json"
	"net/http"

	"github.com/hemu1808/auradeploy/backend/internal/models"
	"github.com/hemu1808/auradeploy/backend/internal/orchestrator"
)

type Handler struct {
	orch orchestrator.Orchestrator
}

func NewHandler(o orchestrator.Orchestrator) *Handler {
	return &Handler{orch: o}
}

// GetApplicationsHandler returns all applications
func (h *Handler) GetApplicationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apps, err := h.orch.GetApplications()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apps)
}

type DeployRequest struct {
	Image    string               `json:"image"`
	Replicas int                  `json:"replicas"`
	EnvVars  []models.EnvVar      `json:"envVars,omitempty"`
	Ports    []models.PortMapping `json:"ports,omitempty"`
}

// DeployApplicationHandler spawns a new service
func (h *Handler) DeployApplicationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Image == "" || req.Replicas < 1 {
		http.Error(w, "Image and valid replica count required", http.StatusBadRequest)
		return
	}

	app, err := h.orch.DeployApplication(req.Image, req.Replicas, req.EnvVars, req.Ports)
	if err != nil {
		if err.Error() == "cannot deploy: not the raft leader" {
			http.Error(w, "Raft leader not elected yet. Please wait and retry.", http.StatusServiceUnavailable)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(app)
}

type ScaleRequest struct {
	Replicas int `json:"replicas"`
}

// ScaleApplicationHandler updates the target replicas
func (h *Handler) ScaleApplicationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	appID := r.URL.Query().Get("id")
	if appID == "" {
		http.Error(w, "Application ID required", http.StatusBadRequest)
		return
	}

	var req ScaleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Replicas < 0 {
		http.Error(w, "Valid replica count required", http.StatusBadRequest)
		return
	}

	app, err := h.orch.ScaleApplication(appID, req.Replicas)
	if err != nil {
		if err.Error() == "cannot scale: not the raft leader" {
			http.Error(w, "Raft leader not elected yet. Please wait and retry.", http.StatusServiceUnavailable)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(app)
}

// RemoveApplicationHandler deletes the service
func (h *Handler) RemoveApplicationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	appID := r.URL.Query().Get("id")
	if appID == "" {
		http.Error(w, "Application ID required", http.StatusBadRequest)
		return
	}

	err := h.orch.RemoveApplication(appID)
	if err != nil {
		if err.Error() == "cannot remove: not the raft leader" {
			http.Error(w, "Raft leader not elected yet. Please wait and retry.", http.StatusServiceUnavailable)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type JoinRequest struct {
	NodeID   string `json:"nodeId"`
	BindAddr string `json:"bindAddr"`
}

// JoinClusterHandler handles requests from other nodes wanting to join the Raft cluster
func (h *Handler) JoinClusterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req JoinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.NodeID == "" || req.BindAddr == "" {
		http.Error(w, "NodeID and BindAddr required", http.StatusBadRequest)
		return
	}

	err := h.orch.JoinCluster(req.NodeID, req.BindAddr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
