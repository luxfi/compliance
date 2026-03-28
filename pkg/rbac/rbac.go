// Package rbac provides role-based access control for compliance modules.
package rbac

import "github.com/luxfi/compliance/pkg/types"

// AllModules returns the canonical list of compliance modules.
// Used by seed data and RBAC role definitions to prevent drift.
func AllModules() []string {
	return []string{
		"kyc", "aml", "applications", "pipelines", "sessions",
		"funds", "esign", "roles", "users", "transactions",
		"reports", "settings", "credentials", "billing", "dashboard",
	}
}

// DefaultRoles returns the standard set of compliance roles.
func DefaultRoles() []*types.Role {
	allModules := AllModules()
	allActions := []string{"read", "write", "delete", "admin"}

	// Owner -- full access to everything.
	ownerPerms := make([]types.Permission, 0, len(allModules)*len(allActions))
	for _, m := range allModules {
		for _, a := range allActions {
			ownerPerms = append(ownerPerms, types.Permission{Module: m, Action: a})
		}
	}

	// Admin -- full access except role management deletion.
	adminPerms := make([]types.Permission, 0)
	for _, m := range allModules {
		for _, a := range allActions {
			if m == "roles" && a == "delete" {
				continue
			}
			adminPerms = append(adminPerms, types.Permission{Module: m, Action: a})
		}
	}

	return []*types.Role{
		{Name: "Owner", Description: "Full access to all modules", Permissions: ownerPerms},
		{Name: "Admin", Description: "Administrative access, cannot delete roles", Permissions: adminPerms},
		{Name: "Manager", Description: "Operational management of onboarding and funds", Permissions: []types.Permission{
			{Module: "kyc", Action: "read"}, {Module: "kyc", Action: "write"},
			{Module: "aml", Action: "read"}, {Module: "aml", Action: "write"},
			{Module: "applications", Action: "read"}, {Module: "applications", Action: "write"},
			{Module: "funds", Action: "read"}, {Module: "funds", Action: "write"},
			{Module: "esign", Action: "read"}, {Module: "esign", Action: "write"},
			{Module: "pipelines", Action: "read"}, {Module: "pipelines", Action: "write"},
			{Module: "sessions", Action: "read"}, {Module: "sessions", Action: "write"},
			{Module: "dashboard", Action: "read"},
			{Module: "transactions", Action: "read"},
		}},
		{Name: "Developer", Description: "Read access for integrations and debugging", Permissions: []types.Permission{
			{Module: "kyc", Action: "read"},
			{Module: "aml", Action: "read"},
			{Module: "applications", Action: "read"},
			{Module: "funds", Action: "read"},
			{Module: "esign", Action: "read"},
			{Module: "pipelines", Action: "read"},
			{Module: "sessions", Action: "read"},
			{Module: "dashboard", Action: "read"},
			{Module: "transactions", Action: "read"},
		}},
		{Name: "Agent", Description: "Investor-facing agent for onboarding sessions", Permissions: []types.Permission{
			{Module: "sessions", Action: "read"}, {Module: "sessions", Action: "write"},
			{Module: "kyc", Action: "read"},
			{Module: "esign", Action: "read"},
			{Module: "dashboard", Action: "read"},
		}},
		{Name: "Reviewer", Description: "Can review and approve applications and AML screenings", Permissions: []types.Permission{
			{Module: "kyc", Action: "read"}, {Module: "kyc", Action: "write"},
			{Module: "aml", Action: "read"}, {Module: "aml", Action: "write"},
			{Module: "applications", Action: "read"}, {Module: "applications", Action: "write"},
			{Module: "dashboard", Action: "read"},
		}},
	}
}

// ComplianceModules returns the static list of compliance modules and their
// available actions, used to build a permission matrix UI.
func ComplianceModules() []types.Module {
	return []types.Module{
		{Name: "kyc", Description: "KYC identity verification", Actions: []string{"read", "write", "admin"}},
		{Name: "aml", Description: "AML screening and monitoring", Actions: []string{"read", "write", "admin"}},
		{Name: "applications", Description: "Onboarding applications", Actions: []string{"read", "write", "admin"}},
		{Name: "funds", Description: "Fund management", Actions: []string{"read", "write", "delete", "admin"}},
		{Name: "esign", Description: "Electronic signatures", Actions: []string{"read", "write", "admin"}},
		{Name: "pipelines", Description: "Onboarding pipelines", Actions: []string{"read", "write", "delete", "admin"}},
		{Name: "sessions", Description: "Investor onboarding sessions", Actions: []string{"read", "write", "admin"}},
		{Name: "roles", Description: "Role-based access control", Actions: []string{"read", "write", "delete", "admin"}},
	}
}

// HasPermission checks if a role has the specified module+action permission.
// The "admin" action on a module implies all other actions.
func HasPermission(role *types.Role, module, action string) bool {
	for _, perm := range role.Permissions {
		if perm.Module == module && perm.Action == action {
			return true
		}
		if perm.Module == module && perm.Action == "admin" {
			return true
		}
	}
	return false
}
