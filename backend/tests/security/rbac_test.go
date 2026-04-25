//go:build security

// RBAC matrix test — every (role, route) cell.
//
// Doctrine refs: Rule 39 (security as core), Rule 44 (testability).
// Plan ADR: docs/adr/0009-rbac-middleware-permission-registry.md.
//
// This is a unit test for `internal/security/rbac.go`; an integration variant
// (S3) wires the middleware into a Fiber app and asserts 403 RFC 7807 on
// permission-deny paths.

package security_test

import (
	"testing"

	"github.com/greenmetrics/backend/internal/security"
)

func TestRoleHasExpectedPermissions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		role     security.Role
		perm     security.Permission
		expected bool
	}{
		// Admin holds every permission.
		{security.RoleAdmin, security.PermMetersRead, true},
		{security.RoleAdmin, security.PermMetersWrite, true},
		{security.RoleAdmin, security.PermReportsGenerate, true},
		{security.RoleAdmin, security.PermAuditRead, true},
		{security.RoleAdmin, security.PermTenantsAdmin, true},
		{security.RoleAdmin, security.PermUsersAdmin, true},

		// Operator holds CRUD on operational data + ingest, but not admin or audit.
		{security.RoleOperator, security.PermMetersWrite, true},
		{security.RoleOperator, security.PermReadingsIngest, true},
		{security.RoleOperator, security.PermPulseIngest, true},
		{security.RoleOperator, security.PermReportsGenerate, true},
		{security.RoleOperator, security.PermTenantsAdmin, false},
		{security.RoleOperator, security.PermUsersAdmin, false},
		{security.RoleOperator, security.PermAuditRead, false},
		{security.RoleOperator, security.PermEmissionFactorsWrite, false},

		// Viewer is read-only.
		{security.RoleViewer, security.PermMetersRead, true},
		{security.RoleViewer, security.PermReadingsRead, true},
		{security.RoleViewer, security.PermReportsRead, true},
		{security.RoleViewer, security.PermAlertsRead, true},
		{security.RoleViewer, security.PermMetersWrite, false},
		{security.RoleViewer, security.PermReadingsIngest, false},
		{security.RoleViewer, security.PermReportsGenerate, false},
		{security.RoleViewer, security.PermAlertsAck, false},
		{security.RoleViewer, security.PermAuditRead, false},

		// Auditor holds read on operational data + audit_log; no writes.
		{security.RoleAuditor, security.PermMetersRead, true},
		{security.RoleAuditor, security.PermAuditRead, true},
		{security.RoleAuditor, security.PermMetersWrite, false},
		{security.RoleAuditor, security.PermReportsGenerate, false},
		{security.RoleAuditor, security.PermTenantsAdmin, false},

		// Unknown role holds nothing.
		{security.Role("unknown"), security.PermMetersRead, false},
	}

	for _, c := range cases {
		c := c
		t.Run(string(c.role)+"/"+string(c.perm), func(t *testing.T) {
			t.Parallel()
			got := security.HasPermission(c.role, c.perm)
			if got != c.expected {
				t.Errorf("HasPermission(%q, %q) = %v; want %v", c.role, c.perm, got, c.expected)
			}
		})
	}
}

func TestPermissionsForUnknownRole(t *testing.T) {
	t.Parallel()
	if got := security.PermissionsFor(security.Role("ghost")); got != nil {
		t.Errorf("PermissionsFor(ghost) = %v; want nil", got)
	}
}

func TestPermissionsForReturnsCopy(t *testing.T) {
	t.Parallel()
	a := security.PermissionsFor(security.RoleViewer)
	b := security.PermissionsFor(security.RoleViewer)
	if len(a) == 0 || len(b) == 0 {
		t.Fatal("expected non-empty viewer permissions")
	}
	a[0] = security.Permission("tampered")
	if b[0] == "tampered" {
		t.Errorf("PermissionsFor returned shared underlying slice — mutation leaked")
	}
}
