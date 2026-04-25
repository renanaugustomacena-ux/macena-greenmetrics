//go:build security

// JWT KID rotation test — overlap window + boot refusal + reject unknown kid.
//
// Doctrine refs: Rule 19, Rule 39, Rule 62.
// Plan ADR: docs/adr/0016-jwt-kid-rotation.md.

package backend_test

import "testing"

func TestKIDOverlapWindowValidatesBothKeys(t *testing.T) {
	t.Skip("scaffold — implement once KID-aware JWTMiddleware lands; tokens signed with kid v3 must validate during overlap")
}

func TestKIDExpiredKeyRejects(t *testing.T) {
	t.Skip("scaffold — token signed with kid v3 must reject after v3.expires_at")
}

func TestUnknownKIDRejectsAlways(t *testing.T) {
	t.Skip("scaffold — token signed with kid not in JWT_KIDS_VALID must reject regardless of secret material")
}

func TestBootRefusesIfCurrentKIDNotInValid(t *testing.T) {
	t.Skip("scaffold — config.Load must refuse if JWT_KID_CURRENT not present in JWT_KIDS_VALID JSON")
}

func TestAlgNoneRejected(t *testing.T) {
	t.Skip("scaffold — token with alg:none header must reject (jwt.WithValidMethods enforces)")
}

func TestNonHS256AlgRejected(t *testing.T) {
	t.Skip("scaffold — token with alg:RS256 must reject")
}
