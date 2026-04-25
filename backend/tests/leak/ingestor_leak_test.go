//go:build leak

// Goroutine leak detection — ingestor lifecycle.
//
// Doctrine refs: Rule 41, Rule 42, Rule 44.
// Tooling: go.uber.org/goleak.

package leak_test

import "testing"

func TestIngestorRunnerNoGoroutineLeak(t *testing.T) {
	t.Skip("scaffold — implement once ingestor refactor lands; defer goleak.VerifyNone(t) at end")

	// Pseudocode:
	//   defer goleak.VerifyNone(t)
	//   ctx, cancel := context.WithCancel(context.Background())
	//   runner := services.NewIngestorRunner(...)
	//   go runner.Start(ctx, sink)
	//   time.Sleep(2 * time.Second)
	//   cancel()
	//   runner.Wait(ctx)   // returns when all loops exit
	//   // goleak.VerifyNone fails if any goroutine still running here
}

func TestPipelineNoLeakOnDrop(t *testing.T) {
	t.Skip("scaffold — saturate pipeline.Submit; verify writer drains + exits cleanly on Stop")
}

func TestWorkerPoolNoLeakOnShutdown(t *testing.T) {
	t.Skip("scaffold — submit work, shutdown, verify no panjf2000/ants goroutines leaked")
}
