//go:build bench

// Bench — pgx CopyFrom throughput target ≥ 50k readings/s on staging.
//
// Doctrine refs: Rule 37, Rule 44.
// Run: go test -bench=. -benchmem ./tests/bench/...

package bench_test

import (
	"testing"
)

func BenchmarkCopyFrom1KBatches(b *testing.B) {
	b.Skip("scaffold — implement when testcontainers fixture lands; populate 1000-row batches via pgx.CopyFrom")
}

func BenchmarkCopyFrom5KBatches(b *testing.B) {
	b.Skip("scaffold — 5000-row batches; expect linear scale")
}

func BenchmarkParallelInsert(b *testing.B) {
	b.Skip("scaffold — N parallel inserter goroutines; verify pool acquire latency stays within budget")
}

func BenchmarkJWTVerify(b *testing.B) {
	b.Skip("scaffold — verify HS256 JWT verify p99 < 20 ms baseline")
}

func BenchmarkValidatorBind(b *testing.B) {
	b.Skip("scaffold — validate.Struct on a representative request body; baseline reflection cost")
}
