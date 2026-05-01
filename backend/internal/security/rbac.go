// Package security — RBAC permission registry + middleware.
//
// Doctrine refs: Rule 19 (security as structural), Rule 39 (backend security as core), Rule 65 (regulated quality).
// Plan ADR: docs/adr/0009-rbac-middleware-permission-registry.md (S3 to author).
// Mitigates: RISK-007 (cross-tenant + role escalation).
//
// Usage in handlers.go:
//
//   protected.Post("/v1/meters",
//       JWTMiddleware(deps),
//       security.RequirePermission(security.PermMetersWrite),
//       meters.Create)
//
// Tests: backend/tests/security/rbac_test.go (table-driven role × route).
package security

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Role is a string-typed RBAC role label. String values must match
// `internal/models/user.go` UserRole constants verbatim so JWT `role` claims
// resolve correctly here.
type Role string

const (
	RoleAdmin    Role = "admin"
	RoleManager  Role = "manager"
	RoleOperator Role = "operator"
	RoleAuditor  Role = "auditor"
	RoleViewer   Role = "viewer"
	RoleReadOnly Role = "readonly"
	// RoleDPO — Data Protection Officer. The only role authorised to invoke
	// GDPR Data Subject Access Request (DSAR) export/erase endpoints. The
	// DPO role is an Italian-flagship requirement (Garante guidance + ARERA
	// 646/2015) and must be assigned to a named human; service accounts may
	// not hold it.
	RoleDPO Role = "dpo"
)

// Permission is a colon-separated `resource:verb` grant.
type Permission string

const (
	PermMetersRead             Permission = "meters:read"
	PermMetersWrite            Permission = "meters:write"
	PermReadingsRead           Permission = "readings:read"
	PermReadingsIngest         Permission = "readings:ingest"
	PermReportsRead            Permission = "reports:read"
	PermReportsGenerate        Permission = "reports:generate"
	PermAlertsRead             Permission = "alerts:read"
	PermAlertsAck              Permission = "alerts:ack"
	PermEmissionFactorsRead    Permission = "emission_factors:read"
	PermEmissionFactorsWrite   Permission = "emission_factors:write"
	PermAuditRead              Permission = "audit:read"
	PermTenantsAdmin           Permission = "tenants:admin"
	PermUsersAdmin             Permission = "users:admin"
	PermJobsRead               Permission = "jobs:read"
	PermPulseIngest            Permission = "pulse:ingest"
	// GDPR DSAR — Data Subject Access Request endpoints. Reserved to the DPO role.
	PermDSARExport             Permission = "dsar:export"
	PermDSARErase              Permission = "dsar:erase"
)

// rolePermissions defines the canonical role → permission map. Modifying this map
// is the only way to change RBAC; per-handler ad-hoc auth is forbidden (Rule 46).
var rolePermissions = map[Role][]Permission{
	RoleAdmin: {
		PermMetersRead, PermMetersWrite,
		PermReadingsRead, PermReadingsIngest,
		PermReportsRead, PermReportsGenerate,
		PermAlertsRead, PermAlertsAck,
		PermEmissionFactorsRead, PermEmissionFactorsWrite,
		PermAuditRead,
		PermTenantsAdmin, PermUsersAdmin,
		PermJobsRead, PermPulseIngest,
	},
	// Manager — admin minus user/tenant administration.
	RoleManager: {
		PermMetersRead, PermMetersWrite,
		PermReadingsRead, PermReadingsIngest,
		PermReportsRead, PermReportsGenerate,
		PermAlertsRead, PermAlertsAck,
		PermEmissionFactorsRead, PermEmissionFactorsWrite,
		PermAuditRead,
		PermJobsRead, PermPulseIngest,
	},
	RoleOperator: {
		PermMetersRead, PermMetersWrite,
		PermReadingsRead, PermReadingsIngest,
		PermReportsRead, PermReportsGenerate,
		PermAlertsRead, PermAlertsAck,
		PermEmissionFactorsRead,
		PermJobsRead, PermPulseIngest,
	},
	RoleAuditor: {
		PermMetersRead,
		PermReadingsRead,
		PermReportsRead,
		PermAlertsRead,
		PermEmissionFactorsRead,
		PermAuditRead,
		PermJobsRead,
	},
	RoleViewer: {
		PermMetersRead,
		PermReadingsRead,
		PermReportsRead,
		PermAlertsRead,
		PermEmissionFactorsRead,
		PermJobsRead,
	},
	// ReadOnly — alias of Viewer for `internal/models/user.go` compatibility.
	RoleReadOnly: {
		PermMetersRead,
		PermReadingsRead,
		PermReportsRead,
		PermAlertsRead,
		PermEmissionFactorsRead,
		PermJobsRead,
	},
	// DPO — Data Protection Officer. ONLY DSAR permissions; DPO does not
	// hold operational read/write rights. Audit log captures every DSAR
	// invocation per Rule 62.
	RoleDPO: {
		PermDSARExport,
		PermDSARErase,
		PermAuditRead,
	},
}

// HasPermission returns true if the role grants the permission.
func HasPermission(role Role, perm Permission) bool {
	perms, ok := rolePermissions[role]
	if !ok {
		return false
	}
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}

// PermissionsFor returns a copy of the permission slice for a role (or nil if unknown).
func PermissionsFor(role Role) []Permission {
	perms, ok := rolePermissions[role]
	if !ok {
		return nil
	}
	out := make([]Permission, len(perms))
	copy(out, perms)
	return out
}

// RequirePermission is a Fiber middleware factory. It reads `c.Locals("user_role")`
// (set by JWTMiddleware), looks the role up in the permission registry, and returns
// 403 RFC 7807 if the permission is not granted.
func RequirePermission(perm Permission) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw, ok := c.Locals("user_role").(string)
		if !ok || strings.TrimSpace(raw) == "" {
			return fiber.NewError(fiber.StatusForbidden, "missing role on token")
		}
		role := Role(raw)
		if !HasPermission(role, perm) {
			return fiber.NewError(fiber.StatusForbidden,
				"role '"+raw+"' lacks permission '"+string(perm)+"'")
		}
		return c.Next()
	}
}

// RequireAny grants if the role holds at least one of the listed permissions.
// Useful where multiple permissions can satisfy an endpoint.
func RequireAny(perms ...Permission) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw, ok := c.Locals("user_role").(string)
		if !ok || strings.TrimSpace(raw) == "" {
			return fiber.NewError(fiber.StatusForbidden, "missing role on token")
		}
		role := Role(raw)
		for _, p := range perms {
			if HasPermission(role, p) {
				return c.Next()
			}
		}
		return fiber.NewError(fiber.StatusForbidden,
			"role '"+raw+"' lacks any required permission")
	}
}
