package models

import "time"

// UserRole enumerates RBAC roles used by GreenMetrics.
type UserRole string

const (
	RoleAdmin     UserRole = "admin"
	RoleManager   UserRole = "manager"
	RoleOperator  UserRole = "operator"
	RoleAuditor   UserRole = "auditor"
	RoleReadOnly  UserRole = "readonly"
)

// User represents an authenticated console / API user.
type User struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         UserRole  `json:"role"`
	FullName     string    `json:"full_name,omitempty"`
	MFAEnabled   bool      `json:"mfa_enabled"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}
