package models

// Role represents a set of permissions
type Role struct {
	ID        string   `json:"id"`
	Verbs     []string `json:"verbs"`     // e.g., ["get", "create", "delete"]
	Resources []string `json:"resources"` // e.g., ["applications", "nodes"]
}

// RoleBinding binds a role to a specific token/user
type RoleBinding struct {
	ID      string `json:"id"`
	RoleID  string `json:"roleId"`
	Subject string `json:"subject"` // e.g., "admin-user" or JWT subject
}

// DefaultAdminRole provides full access
func DefaultAdminRole() Role {
	return Role{
		ID:        "admin-role",
		Verbs:     []string{"*"},
		Resources: []string{"*"},
	}
}
