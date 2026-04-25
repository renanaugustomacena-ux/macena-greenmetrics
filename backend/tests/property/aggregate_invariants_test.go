//go:build property

// Property test — aggregate invariants (CAGGs are sum-consistent).

package property_test

import "testing"

func TestSumOver1DGreaterThanOrEqualToSumOf15MinBuckets(t *testing.T) {
	t.Skip("scaffold — implement with gopter against testcontainer Timescale 16")

	// Property: for any (tenant, meter, period), sum(value over 1d CAGG) ≈
	// sum(value over 15min CAGG) within FP tolerance.
	//
	// Strategy:
	//   1. Spin testcontainer; insert N readings via repository.
	//   2. Refresh CAGGs.
	//   3. Compute both aggregates.
	//   4. Assert difference < 1e-6 (FP tolerance).
}
