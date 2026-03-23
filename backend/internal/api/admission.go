package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/hemu1808/auradeploy/backend/internal/models"
)

// AdmissionController represents a mutating or validating webhook
type AdmissionController func(req *DeployRequest) error

// AdmissionMiddleware acts like a Kubernetes Validating Admission Webhook
// It intercepts Pod/App creation requests and enforces policies (e.g., Pod Security Standards)
func AdmissionMiddleware(next http.HandlerFunc, controllers ...AdmissionController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r) // Only intercept creations
			return
		}

		// Read the body into a buffer so we can decode it, then put it back
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		var req DeployRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Run all validating/mutating controllers
		for _, ctrl := range controllers {
			if err := ctrl(&req); err != nil {
				http.Error(w, "Admission Webhook Denied: "+err.Error(), http.StatusNotAcceptable)
				return
			}
		}

		// Restore the body for the actual handler
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		next.ServeHTTP(w, r)
	}
}

// PodSecurityStandard implements a baseline security check
func PodSecurityStandard(req *DeployRequest) error {
	for _, env := range req.EnvVars {
		if env.Key == "RUN_AS_ROOT" && env.Value == "true" {
			// In standard K8s this is checked via SecurityContext structs.
			// Here we simulate it by blocking specific env flags or privileged ports
			return models.ErrPrivilegedContainer
		}
	}
	
	// Disallow binding to privileged host ports
	for _, port := range req.Ports {
		if port.HostPort < 1024 {
			return models.ErrPrivilegedPort
		}
	}

	return nil
}
