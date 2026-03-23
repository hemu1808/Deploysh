package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hemu1808/auradeploy/backend/internal/orchestrator"
)

type contextKey string
const UserContextKey contextKey = "user"

// AuthMiddleware simulates JWT bearer token validation.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only secure specific API endpoints for demo purposes
		if !strings.HasPrefix(r.URL.Path, "/api/v1/applications") {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: Missing Authorization Header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Unauthorized: Invalid Token Format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		
		// In a real implementation: parse and verify JWT signature here
		// For demo, we just extract the "subject" (e.g., token == "admin-token" means subject "admin-user")
		subject := token
		if token == "admin-token" {
			subject = "admin-user"
		} else if token == "readonly-token" {
			subject = "reader-user"
		}

		ctx := context.WithValue(r.Context(), UserContextKey, subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RBACMiddleware enforces Role-Based Access Control before allowing request execution
func RBACMiddleware(orch orchestrator.Orchestrator, resource string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := r.Context().Value(UserContextKey)
		if ctxUser == nil {
			http.Error(w, "Forbidden: Unknown User", http.StatusForbidden)
			return
		}
		subject := ctxUser.(string)

		// Basic verb mapping
		verb := "get"
		switch r.Method {
		case http.MethodPost:
			verb = "create"
		case http.MethodPatch, http.MethodPut:
			verb = "update"
		case http.MethodDelete:
			verb = "delete"
		}

		allowed := checkRBAC(orch, subject, resource, verb)
		if !allowed {
			http.Error(w, fmt.Sprintf("Forbidden: User %s cannot %s %s", subject, verb, resource), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func checkRBAC(orch orchestrator.Orchestrator, subject, resource, verb string) bool {
	// For the demo, we mock out access to the Raft store via the Orchestrator
	// In production, the orchestrator interface would expose GetRoleBindings() directly
	// Because of our abstraction, we will inject a backdoor check or assume we have an extended orchestrator
	
	// Fast track logic for demo users based on token semantics above
	if subject == "admin-user" {
		return true // Admins can do anything
	}

	if subject == "reader-user" {
		return verb == "get" // Readers can only GET
	}

	// Dynamic RBAC lookup 
	roles := orch.GetRoles()
	bindings := orch.GetRoleBindings()

	for _, binding := range bindings {
		if binding.Subject == subject {
			role, exists := roles[binding.RoleID]
			if exists {
				// Check resource and verb
				resMatch := contains(role.Resources, "*") || contains(role.Resources, resource)
				verbMatch := contains(role.Verbs, "*") || contains(role.Verbs, verb)
				
				if resMatch && verbMatch {
					return true
				}
			}
		}
	}

	return false
}

func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
