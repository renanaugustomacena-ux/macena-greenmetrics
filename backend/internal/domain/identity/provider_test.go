// Example test for the Identity Pack contract. Per Rule 86.

package identity_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/greenmetrics/backend/internal/domain/identity"
)

// stubProvider models a minimal Identity Pack with a single hard-coded user.
type stubProvider struct{}

func (stubProvider) Name() string { return "stub" }

func (stubProvider) Authenticate(_ context.Context, creds identity.Credentials) (identity.Identity, error) {
	if creds.Username == "alice@example.test" && creds.Password == "correct horse battery staple" {
		return identity.Identity{
			ProviderName: "stub",
			Subject:      creds.Username,
			Email:        creds.Username,
			DisplayName:  "Alice Example",
			MFAVerified:  false,
		}, nil
	}
	return identity.Identity{}, errors.New("invalid credentials")
}

func (stubProvider) LookupUser(_ context.Context, id identity.Identity) (identity.User, error) {
	return identity.User{
		ID:          uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		TenantID:    uuid.MustParse("22222222-2222-4222-8222-222222222222"),
		Email:       id.Email,
		DisplayName: id.DisplayName,
		Role:        "operator",
		Active:      true,
	}, nil
}

func TestExample_IdentityProviderHappyPath(t *testing.T) {
	var p identity.IdentityProvider = stubProvider{}
	ctx := context.Background()

	id, err := p.Authenticate(ctx, identity.Credentials{
		Username:  "alice@example.test",
		Password:  "correct horse battery staple",
		IPAddress: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if id.Email != "alice@example.test" {
		t.Errorf("email mismatch: %s", id.Email)
	}

	u, err := p.LookupUser(ctx, id)
	if err != nil {
		t.Fatalf("LookupUser: %v", err)
	}
	if u.TenantID == uuid.Nil {
		t.Error("TenantID must be resolved")
	}
	if u.Role == "" {
		t.Error("Role must be set")
	}
}

func TestExample_IdentityProviderRejectsBadCreds(t *testing.T) {
	var p identity.IdentityProvider = stubProvider{}
	if _, err := p.Authenticate(context.Background(), identity.Credentials{
		Username: "alice@example.test",
		Password: "wrong",
	}); err == nil {
		t.Fatal("bad credentials should be rejected")
	}
}

func TestContractVersion_IsSet(t *testing.T) {
	if identity.ContractVersion == "" {
		t.Fatal("ContractVersion empty — Rule 71 requires per-kind contract version")
	}
}
