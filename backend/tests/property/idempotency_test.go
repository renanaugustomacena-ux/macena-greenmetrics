//go:build property

// Property test — idempotency replay invariance.
//
// Doctrine refs: Rule 35, Rule 44.
// Plan ADR: docs/adr/0011-postgres-rls-defence-in-depth.md (cross-ref).

package property_test

import (
	"testing"
)

func TestIdempotencyReplayInvariance(t *testing.T) {
	t.Skip("scaffold — implement with gopter when integration fixture lands; depends on tests/integration/postgres_setup.go")

	// Property: for any (tenant, key, body), invoking the API twice with the same
	// Idempotency-Key + same body yields identical response and no duplicate state.
	//
	// Strategy:
	//   1. Spin testcontainer Timescale 16 + apply migrations.
	//   2. Boot in-process Fiber + IdempotencyMiddleware + IdempotencyStore.
	//   3. gopter.QuickCheck:
	//      - generate (key, payload) pairs.
	//      - call POST /v1/readings/ingest twice.
	//      - assert: response status + body identical; row count unchanged on second call.
	//   4. Bonus: replay with different body returns 422 Conflict.
}
