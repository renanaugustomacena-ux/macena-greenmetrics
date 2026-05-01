// no_hardcoded_brand_test.go — white-label conformance test.
//
// Doctrine refs: Rule 80 (white-label conformance), Rule 82 (engagement
// overlay supersedes Core defaults).
// Charter ref: §6 (white-label readiness), §11 (single-tenant by default
// — engagement forks brand without forking the codebase).
//
// This test enforces that an explicit list of "must-be-clean" files
// contains NO hard-coded brand strings (product name, vendor name).
// Engagement-overridable surfaces — UI components, email templates, PDF
// renderers — read from `config/branding.yaml` via `frontend/src/lib/
// branding.ts` (frontend) or the equivalent backend reader (Phase F).
//
// The "must-be-clean" list grows as white-label migration PRs land. New
// surfaces are added here together with the migration that cleans them.
//
// Allowed surfaces (NOT checked) — template-author identity, regulatory-
// citation, and changelog surfaces are exempt by design:
//   - docs/                      (template author voice; not tenant-visible)
//   - packs/*/CHARTER.md         (regulatory-citation cross-refs)
//   - CHANGELOG.md, AUTHORS, LICENSE, NOTICE, COPYRIGHT, TRADEMARK*, README*
//   - .github/*, devcontainer    (template-internal)
//   - frontend/src/lib/branding.* (the brand source-of-truth itself)
//   - *.test.ts, *_test.go, tests/* (test fixtures)
//
// Build tag `conformance` so the test runs in CI's conformance lane only,
// not in the default `go test ./...`.

//go:build conformance
// +build conformance

package conformance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// brandStrings is the set of hard-coded tokens that may NOT appear in
// must-be-clean files. Engagement forks adding new product names extend
// this list; the test fails if any of these appears in a checked surface.
var brandStrings = []string{
	"GreenMetrics",
	"greenmetrics",
}

// mustBeClean is the explicit allowlist of files that MUST contain zero
// hard-coded brand tokens. Each entry is a repository-relative path. The
// list grows as white-label migration PRs clean additional surfaces.
//
// Phase E Sprint S7 (this PR): cleaned the SvelteKit root layout and the
// HTML shell (description meta moved into <svelte:head> driven by
// `$lib/branding`). Subsequent PRs add reports/, meters/, carbon/, and
// the home page.
var mustBeClean = []string{
	"frontend/src/routes/+layout.svelte",
	"frontend/src/app.html",
}

func TestNoHardcodedBrandInCleanedSurfaces(t *testing.T) {
	repoRoot := findRepoRoot(t)
	for _, rel := range mustBeClean {
		full := filepath.Join(repoRoot, rel)
		body, err := os.ReadFile(full)
		if err != nil {
			t.Errorf("must-be-clean file %s: %v", rel, err)
			continue
		}
		text := string(body)
		for _, brand := range brandStrings {
			if strings.Contains(text, brand) {
				t.Errorf("must-be-clean file %s contains hard-coded brand %q — read from $lib/branding instead",
					rel, brand)
			}
		}
	}
}

// TestBrandingSourceFileExists guards against accidental deletion of the
// branding source-of-truth — engagements depend on its presence.
func TestBrandingSourceFileExists(t *testing.T) {
	repoRoot := findRepoRoot(t)
	for _, rel := range []string{
		"config/branding.yaml",
		"frontend/scripts/gen-branding.mjs",
		"frontend/src/lib/branding.ts",
	} {
		full := filepath.Join(repoRoot, rel)
		if _, err := os.Stat(full); err != nil {
			t.Errorf("required white-label source file %s missing: %v", rel, err)
		}
	}
}

// findRepoRoot walks up from the test file looking for the .git directory.
// (Conformance tests run from the package dir under backend/tests/conformance/,
// so we need to walk up to the repo root to inspect frontend/ + config/.)
func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("repository root (.git) not found within 8 ancestor directories of test cwd")
	return ""
}
