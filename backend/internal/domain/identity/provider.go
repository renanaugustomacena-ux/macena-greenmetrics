// Package identity defines the Pack contract for authentication providers.
//
// Doctrine refs: Rules 70 (Pack manifests itself), 87 (Pack acceptance),
// 170 (JWT pinned + KID rotated), 181 (TOTP MFA admin), 182 (lockout).
// Charter ref: §3.2 Identity Packs. ADR-0023 records the interface adoption.
//
// An Identity Pack at packs/identity/<id>/ implements the IdentityProvider
// interface below. The default Identity Pack is the local-DB authenticator
// (bcrypt+pepper, IP+email lockout, TOTP for admin). Phase E Sprint S8
// adds SAML 2.0 and OIDC providers. Future: SPID, CIE.
//
// Identity Packs do NOT issue JWTs themselves — the JWT issuance lives in
// Core's `internal/security/jwt.go` per Rule 170. The Identity Pack proves
// who the user is; Core decides what to put in the token.
package identity

import (
	"context"

	"github.com/google/uuid"
)

// ContractVersion is the SemVer of this Pack-contract package. Per Rule 71.
const ContractVersion = "1.0.0"

// Credentials carries the authentication input. The shape is intentionally
// permissive — different providers consume different fields. Concrete
// providers ignore fields they don't use.
//
// For local-DB: Username + Password (and optional TOTP).
// For SAML 2.0: SAMLResponse (raw assertion) + RelayState.
// For OIDC: Code + RedirectURI (RP-managed flow) or IDToken (already-validated).
type Credentials struct {
	// Common fields.
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	TOTP     string `json:"totp,omitempty"`

	// SAML 2.0.
	SAMLResponse string `json:"saml_response,omitempty"`
	RelayState   string `json:"relay_state,omitempty"`

	// OIDC.
	Code        string `json:"code,omitempty"`
	RedirectURI string `json:"redirect_uri,omitempty"`
	IDToken     string `json:"id_token,omitempty"`

	// IP for lockout / audit (set by middleware, not by the user).
	IPAddress string `json:"-"`
}

// Identity is the result of a successful Authenticate. Core uses it to
// look up or just-in-time-provision the User and to issue a JWT.
type Identity struct {
	ProviderName string            `json:"provider"`
	Subject      string            `json:"subject"`
	Email        string            `json:"email,omitempty"`
	DisplayName  string            `json:"display_name,omitempty"`
	Attributes   map[string]string `json:"attributes,omitempty"`
	MFAVerified  bool              `json:"mfa_verified"`
}

// User is the local user record post-authentication. Identity Packs that
// don't already know the User's tenant (e.g. SAML JIT) return a User with
// TenantID = uuid.Nil; Core resolves the tenant from the JIT-mapping rules.
type User struct {
	ID          uuid.UUID `json:"user_id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Role        string    `json:"role"`
	Active      bool      `json:"active"`
}

// IdentityProvider is the Pack-contract for authentication providers.
type IdentityProvider interface {
	// Name is the provider identifier (matches Pack id; e.g. "local_db",
	// "saml", "oidc", "spid", "cie").
	Name() string

	// Authenticate verifies the credentials and returns the proven Identity.
	// Authentication failures return a typed error that Core's middleware
	// maps to HTTP 401 + Problem Details. Lockout failures return a typed
	// error that maps to HTTP 429.
	Authenticate(ctx context.Context, creds Credentials) (Identity, error)

	// LookupUser resolves an Identity to a local User record. JIT
	// provisioning happens here for providers that support it (SAML, OIDC).
	// Local-DB providers simply read from the users table.
	LookupUser(ctx context.Context, identity Identity) (User, error)
}
