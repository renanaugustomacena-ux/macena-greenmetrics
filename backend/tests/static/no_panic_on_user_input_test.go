//go:build static

// Static check — no `panic(` in internal/handlers/*.go.
//
// Doctrine refs: Rule 46 (REJ-15).
// CI: `grep -rn 'panic(' internal/handlers/` must return 0.

package static_test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestNoPanicInHandlers(t *testing.T) {
	cmd := exec.Command("grep", "-rn", "panic(", "../../internal/handlers")
	out, _ := cmd.CombinedOutput()
	hits := strings.TrimSpace(string(out))
	if hits != "" {
		t.Errorf("panic() in handlers (REJ-15):\n%s", hits)
	}
}
