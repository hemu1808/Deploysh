package api

import (
	"io"
	"net/http"

	"github.com/hemu1808/auradeploy/backend/internal/gitops"
)

type GitOpsHandler struct {
	reconciler *gitops.Reconciler
}

func NewGitOpsHandler(r *gitops.Reconciler) *GitOpsHandler {
	return &GitOpsHandler{reconciler: r}
}

// WebhookHandler receives a YAML body and applies it
func (h *GitOpsHandler) WebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	apps, err := gitops.ParseManifests(data)
	if err != nil {
		http.Error(w, "Failed to parse YAML manifest: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.reconciler.Reconcile(apps); err != nil {
		http.Error(w, "Reconciliation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("GitOps Deployment Successful\n"))
}
